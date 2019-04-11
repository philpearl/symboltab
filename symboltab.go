// Package symboltab is a symbol table. It converts strings to sequence numbers. This is useful
// for things like graph algorithms, where IDs are stored and compared a lot.
//
// symboltab is optimised for storing a lot of strings, so things are optimised for reducing
// work for the GC
package symboltab

import (
	"math/bits"

	"github.com/philpearl/aeshash"
	"github.com/philpearl/stringbank"
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
	return &SymbolTab{
		table: table{
			hashes:   make([]uint32, cap),
			sequence: make([]int32, cap),
		},
	}
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

	// We're going to add to the table, make sure it is big enough
	i.resize()

	if i.oldTable.len() != 0 {
		// If we're resizing currently, then do some resizing work
		i.resizeWork()

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
				return cursor, table.sequence[cursor]
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
	for k := 0; k < 16; k++ {
		offset := k + i.oldTableCursor
		if seq := i.oldTable.sequence[offset]; seq != 0 {
			i.copyEntryToTable(i.table, i.oldTable.hashes[offset], i.oldTable.sequence[offset])
			// The entry can exist in the old and new versions of the table without
			// problems. If we did try to delete from the old table we'd have issues
			// searching forward from clashing entries.
		}
	}
	i.oldTableCursor += 16
	if i.oldTableCursor >= l {
		// resizing is complete - clear out the old table
		i.oldTable.hashes = nil
		i.oldTable.sequence = nil
		i.oldTableCursor = 0
	}
}

func (i *SymbolTab) resize() {
	if i.table.hashes == nil {
		// Makes zero value of SymbolTab useful
		i.table.hashes = make([]uint32, 16)
		i.table.sequence = make([]int32, 16)
	}

	if i.count < i.table.len()/loadFactor {
		// Not full enough to grow the table
		return
	}

	if i.oldTable.hashes == nil {
		// Not already resizing, so kick off the process. Note that despite all the work we do to try to be
		// clever, just allocating these slices can cause a considerable amount of work, presumably because
		// they are set to zero.
		i.oldTable, i.table = i.table, table{
			hashes:   make([]uint32, len(i.table.hashes)*2),
			sequence: make([]int32, len(i.table.sequence)*2),
		}
	}
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

func (t table) len() int {
	return len(t.hashes)
}
