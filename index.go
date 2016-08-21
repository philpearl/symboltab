package symboltab

import (
	"fmt"
	"unsafe"
)

// indexEntry is part of an array indexed by an int32, and contains an offset &
// length that point to a string
type indexEntry struct {
	offset int32
	length int32
}

const sizeofIndexEntry = 8

type index mMapFile

func openIndex(baseName string) (*index, error) {
	filename := fmt.Sprintf("%s.idx", baseName)
	mMap, err := openMmapFile(filename)

	return (*index)(mMap), err
}

// Get retrieves the i'th offset and length from the index
func (idx *index) Get(i int) (offset, length int32) {
	entry := (*indexEntry)(unsafe.Pointer(&idx.mMapData[i*sizeofIndexEntry]))
	return entry.offset, entry.length
}

// Len returns the number of items in the index
func (idx *index) Len() int {
	return len(idx.mMapData) / sizeofIndexEntry
}

// Put appends an offset and length to the index
func (idx *index) Put(offset, length int32) error {
	entry := &indexEntry{offset: offset, length: length}
	array := *(*[sizeofIndexEntry]byte)(unsafe.Pointer(entry))

	return (*mMapFile)(idx).append(array[:])
}
