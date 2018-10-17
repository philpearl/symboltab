package symboltab

const intbanksize = 1 << 9

type intbank struct {
	slabs [][]int
}

func (ib *intbank) save(sequence int32, offset int) {
	sequence-- // externally sequence starts at 1
	slabNo := int(sequence / intbanksize)
	slabOffset := int(sequence % intbanksize)

	for len(ib.slabs) <= slabNo {
		ib.slabs = append(ib.slabs, make([]int, intbanksize))
	}

	ib.slabs[slabNo][slabOffset] = offset
}

func (ib *intbank) lookup(sequence int32) int {
	sequence-- // externally, sequence starts at 1
	slabNo := int(sequence / intbanksize)
	slabOffset := int(sequence % intbanksize)

	return ib.slabs[slabNo][slabOffset]
}
