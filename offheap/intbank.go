package offheap

import (
	"github.com/philpearl/mmap"
)

const intbanksize = 1 << 12

type intbank struct {
	slabs [][]int
}

func (ib *intbank) close() {
	for _, s := range ib.slabs {
		mmap.Free(s)
	}
	ib.slabs = nil
}

func (ib *intbank) save(sequence uint32, offset int) {
	sequence-- // externally sequence starts at 1
	slabNo := int(sequence / intbanksize)
	slabOffset := int(sequence % intbanksize)

	for len(ib.slabs) <= slabNo {
		ns, _ := mmap.Alloc[int](intbanksize)
		ib.slabs = append(ib.slabs, ns)
	}

	ib.slabs[slabNo][slabOffset] = offset
}

func (ib *intbank) lookup(sequence uint32) int {
	sequence-- // externally, sequence starts at 1
	slabNo := int(sequence / intbanksize)
	slabOffset := int(sequence % intbanksize)

	return ib.slabs[slabNo][slabOffset]
}
