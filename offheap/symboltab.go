// Package offheap is an off-heap symbol table. It converts strings to sequence numbers. This is useful
// for things like graph algorithms, where IDs are stored and compared a lot.
//
// symboltab is optimised for storing a lot of strings, so things are optimised for reducing
// work for the GC
package offheap

import (
	"math/bits"
	"reflect"
	"unsafe"

	"github.com/philpearl/aeshash"
	"github.com/philpearl/mmap"
	stringbank "github.com/philpearl/stringbank/offheap"
)

// Our space costs are 8 bytes per entry. With a load factor of 0.5 (written as 2 here for reasons) that's
// increased to at least 16 bytes per entry
const loadFactor = 2

// SymbolTab is the symbol table. Allocate it via New()
type SymbolTab struct {
	sb             stringbank.Stringbank
	table          table
	oldTable       table
	count          int
	oldTableCursor int
	ib             intbank
}

// New creates a new SymbolTab. cap is the initial capacity of the table - it will grow
// automatically when needed
func New(cap int) *SymbolTab {
	// want to allocate a table large enough to hold cap without growing
	cap = cap * loadFactor
	if cap < 16 {
		cap = 16
	} else {
		cap = 1 << uint(64-bits.LeadingZeros(uint(cap-1)))
	}
	var t table
	t.init(cap)
	return &SymbolTab{
		table: t,
	}
}

// Close releases resources associated with the SymbolTab
func (i *SymbolTab) Close() {
	i.sb.Close()
	i.table.close()
	i.oldTable.close()
	i.oldTableCursor = 0
	i.count = 0
	i.ib.close()
}

// Len returns the number of unique strings stored
func (i *SymbolTab) Len() int {
	return i.count
}

// Cap returns the size of the SymbolTab table
func (i *SymbolTab) Cap() int {
	return i.table.len()
}

// SymbolSize contains the approximate size of string storage in the symboltable. This will be an over-estimate and
// includes as yet unused and wasted space
func (i *SymbolTab) SymbolSize() int {
	return i.sb.Size()
}

// SequenceToString looks up a string by its sequence number. Obtain the sequence number
// for a string with StringToSequence
func (i *SymbolTab) SequenceToString(seq int32) string {
	// Look up the stringbank offset for this sequence number, then get the string
	offset := i.ib.lookup(seq)
	return i.sb.Get(offset)
}

// StringToSequence looks up the string val and returns its sequence number seq. If val does
// not currently exist in the symbol table, it will add it if addNew is true. found indicates
// whether val was already present in the SymbolTab
func (i *SymbolTab) StringToSequence(val string, addNew bool) (seq int32, found bool) {
	// we use a hashtable where the keys are stringbank offsets, but comparisons are done on
	// strings. There is no value to store
	hash := aeshash.Hash(val)

	if addNew {
		// We're going to add to the table, make sure it is big enough
		// We make sure we don't do any resizing work if we're not writing data as it will surprise folk who
		// might hold just a read lock while reading.
		i.resize()
	}

	if i.oldTable.len() != 0 {
		if addNew {
			// If we're resizing currently, then do some resizing work
			i.resizeWork()
		}

		// The data might still be only in the old table, so look there first. If we find the
		// data here then we can just go with that answer. But if not it may be in the new table
		// only. Certainly if we add we want to add to the new table
		_, sequence := i.findInTable(i.oldTable, val, hash)
		if sequence != 0 {
			return sequence, true
		}
	}

	cursor, sequence := i.findInTable(i.table, val, hash)
	if sequence != 0 {
		return sequence, true
	}

	if !addNew {
		return 0, false
	}

	// String was not found, so we want to store it. Cursor is the index where we should
	// store it
	i.count++
	sequence = int32(i.count)
	i.table.hashes[cursor] = hash
	i.table.sequence[cursor] = sequence

	offset := i.sb.Save(val)
	i.ib.save(sequence, offset)

	return sequence, false
}

// findInTable find the string val in the hash table. If the string is present, it returns the
// place in the table where it was found, plus the stringbank offset of the string + 1
func (i *SymbolTab) findInTable(table table, val string, hashVal uint32) (cursor int, sequence int32) {
	l := table.len()
	if l == 0 {
		return 0, 0
	}
	cursor = int(hashVal) & (l - 1)
	start := cursor
	for table.sequence[cursor] != 0 {
		if table.hashes[cursor] == hashVal {
			if seq := table.sequence[cursor]; i.sb.Get(int(i.ib.lookup(seq))) == val {
				return cursor, seq
			}
		}
		cursor++
		if cursor == l {
			cursor = 0
		}
		if cursor == start {
			panic("out of space!")
		}
	}
	return cursor, 0
}

func (i *SymbolTab) copyEntryToTable(table table, hash uint32, seq int32) {
	l := table.len()
	cursor := int(hash) & (l - 1)
	start := cursor
	for table.sequence[cursor] != 0 {
		// the entry we're copying in is guaranteed not to be already
		// present, so we're just looking for an empty space
		cursor++
		if cursor == l {
			cursor = 0
		}
		if cursor == start {
			panic("out of space (resize)!")
		}
	}
	table.hashes[cursor] = hash
	table.sequence[cursor] = seq
}

func (i *SymbolTab) resizeWork() {
	// We copy items between tables 16 at a time. Since we do this every time
	// anyone writes to the table we won't run out of space in the new table
	// before this is complete
	l := i.oldTable.len()
	if l == 0 {
		return
	}
	// original size is 16, and we double to create new tables, so size should always be a multiple of 16
	for k, seq := range i.oldTable.sequence[i.oldTableCursor : i.oldTableCursor+16] {
		if seq != 0 {
			offset := k + i.oldTableCursor
			i.copyEntryToTable(i.table, i.oldTable.hashes[offset], seq)
			// The entry can exist in the old and new versions of the table without
			// problems. If we did try to delete from the old table we'd have issues
			// searching forward from clashing entries.
		}
	}
	i.oldTableCursor += 16
	if i.oldTableCursor >= l {
		// resizing is complete - clear out the old table
		i.oldTable.close()
		i.oldTableCursor = 0
	}
}

func (i *SymbolTab) resize() {
	if i.table.hashes == nil {
		// Makes zero value of SymbolTab useful
		i.table.init(16)
	}

	if i.count < i.table.len()/loadFactor {
		// Not full enough to grow the table
		return
	}

	if i.oldTable.hashes == nil {
		// Not already resizing, so kick off the process. Note that despite all the work we do to try to be
		// clever, just allocating these slices can cause a considerable amount of work, presumably because
		// they are set to zero.
		var newTable table
		newTable.init(i.table.len() * 2)
		i.oldTable, i.table = i.table, newTable
	}
}

func makeUint32Slice(size int) []uint32 {
	slice, _ := mmap.Alloc(unsafe.Sizeof(uint32(0)), size)
	slice.Len = size
	return *(*[]uint32)(unsafe.Pointer(&slice))
}

func makeInt32Slice(size int) []int32 {
	slice, _ := mmap.Alloc(unsafe.Sizeof(int32(0)), size)
	slice.Len = size
	return *(*[]int32)(unsafe.Pointer(&slice))
}

// table represents a hash table. We keep the strings and hashes separate in
// case we want to use different size types in the future
type table struct {
	// We keep hashes in the table to speed up resizing, and also stepping through
	// entries that have different hashes but hit the same bucket
	hashes []uint32
	// sequence contains the sequence numbers of the entries
	sequence []int32
}

func (t *table) init(cap int) {
	t.hashes = makeUint32Slice(cap)
	t.sequence = makeInt32Slice(cap)
}

func (t table) len() int {
	return len(t.hashes)
}

func (t *table) close() {
	if t.hashes != nil {
		mmap.Free(*(*reflect.SliceHeader)(unsafe.Pointer(&t.hashes)), unsafe.Sizeof(uint32(0)))
		t.hashes = nil
	}
	if t.sequence != nil {
		mmap.Free(*(*reflect.SliceHeader)(unsafe.Pointer(&t.sequence)), unsafe.Sizeof(int32(0)))
		t.sequence = nil
	}
}
