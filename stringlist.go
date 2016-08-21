package symboltab

import (
	"fmt"
)

// stringlist is a memory mapped list of strings. Strings are just added to the
// end of the memory-mapped file, and so an external index of offsets and lengths
// is required to read the strings out again
type stringList mMapFile

func openStringList(baseName string) (*stringList, error) {
	filename := fmt.Sprintf("%s.sl", baseName)
	mMap, err := openMmapFile(filename)

	return (*stringList)(mMap), err
}

// Get returns a string from the string list. Well, we use []byte. But it is
// kind of the same thing, and a bit more general.
func (sl *stringList) Get(offset, length int32) []byte {
	return sl.mMapData[offset : offset+length]
}

// Append adds a new string to the string list. It returns the offset of the
// string or any error encountered
func (sl *stringList) Append(data []byte) (offset int32, err error) {
	// First implementation simply writes the data immediately
	offset = int32(len(sl.mMapData))
	err = (*mMapFile)(sl).append(data)
	return
}
