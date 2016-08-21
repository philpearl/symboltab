package symboltab

import (
	"fmt"
	"os"
	"syscall"
)

type mMapFile struct {
	file     *os.File
	mMapData []byte
}

func openMmapFile(filename string) (*mMapFile, error) {
	f, err := os.OpenFile(filename, os.O_RDWR|os.O_APPEND|os.O_CREATE, 0600)
	if err != nil {
		return nil, err
	}

	fi, err := f.Stat()
	if err != nil {
		return nil, err
	}
	var mMapData []byte
	if fi.Size() > 0 {
		mMapData, err = syscall.Mmap(int(f.Fd()), 0, int(fi.Size()), syscall.PROT_READ, syscall.MAP_SHARED)
		if err != nil {
			return nil, fmt.Errorf("Failed to map file. %v", err)
		}
	}
	return &mMapFile{file: f, mMapData: mMapData}, nil
}

// append adds bytes to the end of the memory mapped file, then remaps it to
// extend the length of the mapping to include the data added.
func (m *mMapFile) append(data []byte) error {
	n, err := m.file.Write(data)
	if err != nil {
		return err
	}

	if n < len(data) {
		return fmt.Errorf("Not enough bytes written. Wrote %d, expected %d", n, len(data))
	}

	length := len(m.mMapData)
	if m.mMapData != nil {
		if err := syscall.Munmap(m.mMapData); err != nil {
			return err
		}
	}

	mMapData, err := syscall.Mmap(int(m.file.Fd()), 0, n+length, syscall.PROT_READ, syscall.MAP_SHARED)
	if err != nil {
		return err
	}
	m.mMapData = mMapData
	return nil
}
