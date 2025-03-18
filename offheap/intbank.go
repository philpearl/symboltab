package offheap

import (
	"reflect"
	"unsafe"

	"github.com/philpearl/mmap"
)

const intbanksize = 1 << 12

type intbank struct {
	slabs [][]int
}

func (ib *intbank) close() {
	for _, s := range ib.slabs {
		mmap.Free(*(*reflect.SliceHeader)(unsafe.Pointer(&s)), unsafe.Sizeof(int(0)))
	}
	ib.slabs = nil
}

func (ib *intbank) save(sequence uint32, offset int) {
	sequence-- // externally sequence starts at 1
	slabNo := int(sequence / intbanksize)
	slabOffset := int(sequence % intbanksize)

	for len(ib.slabs) <= slabNo {
		ns, _ := mmap.Alloc(unsafe.Sizeof(int(0)), intbanksize)
		ns.Len = intbanksize
		ib.slabs = append(ib.slabs, *(*[]int)(unsafe.Pointer(&ns)))
	}

	ib.slabs[slabNo][slabOffset] = offset
}

func (ib *intbank) lookup(sequence uint32) int {
	sequence-- // externally, sequence starts at 1
	slabNo := int(sequence / intbanksize)
	slabOffset := int(sequence % intbanksize)

	return ib.slabs[slabNo][slabOffset]
}
