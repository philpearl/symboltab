/*

A symbol table maps strings to integers & vice-versa. This implementation does
only the reverse part - maps integers to strings. The integers are allocated
from 0 sequentially.

When a new string is added it is added to the end of the string list. The bytes
are simply added to the end of the list. The offset in the string list is written
into an index entry, as is the length of the string.

We have separate memory-mapped files for the string-list and the index. These
can then just grow as required.

Reading the data is done via the memory-mapped file. So just like things are
in memory.

When we write data we append to the files, then remap so the memory-map covers
the new data.


Note this code lacks locks (so it isn't process or thread safe), and the reversed
strings may become invalid whenever you write to the table as the memory maps may
move to another location
*/

package symboltab

import (
	"fmt"
	"os"
)

type SymbolTable interface {
	Add(symbol string) (int32, error)
	Reverse(index int32) string
}

type symbolTableImpl struct {
	index      *index
	stringList *stringList
}

func OpenSymbolTable(name string) (SymbolTable, error) {
	idx, err := openIndex(name)
	if err != nil {
		return nil, err
	}
	sl, err := openStringList(name)
	if err != nil {
		return nil, err
	}

	return &symbolTableImpl{
		index:      idx,
		stringList: sl,
	}, nil
}

func DeleteSymbolTable(name string) error {
	if err := os.Remove(fmt.Sprintf("%s.sl", name)); err != nil {
		return err
	}
	return os.Remove(fmt.Sprintf("%s.idx", name))
}

// Add adds a symbol to the table. Returns the index into the table.
//
// Does not currently implement symbol lookup to ensure uniqueness
func (st *symbolTableImpl) Add(symbol string) (int32, error) {
	// Note we only implement the reverse symbol table here.

	index := st.index.Len()

	offset, err := st.stringList.Append([]byte(symbol))
	if err != nil {
		return 0, err
	}

	return int32(index), st.index.Put(offset, int32(len(symbol)))
}

// Reverse finds you a symbol for an index. BEWARE the string may become invalid
// next time anyone writes to the table.
func (st *symbolTableImpl) Reverse(index int32) string {
	offset, length := st.index.Get(int(index))

	data := st.stringList.Get(offset, length)
	return string(data)
}
