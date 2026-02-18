package offheap

import (
	"encoding/binary"
	"fmt"
	"io"
	"unsafe"

	"github.com/philpearl/mmap"
)

const intbanksize = 1 << 12

type intbank struct {
	slabs []*[intbanksize]int
}

func (ib *intbank) close() {
	for _, s := range ib.slabs {
		mmap.Free((*s)[:])
	}
	ib.slabs = nil
}

func (ib *intbank) save(sequence uint32, offset int) {
	sequence-- // externally sequence starts at 1
	slabNo := int(sequence / intbanksize)
	slabOffset := int(sequence % intbanksize)

	for len(ib.slabs) <= slabNo {
		ns, _ := mmap.Alloc[int](intbanksize)
		ib.slabs = append(ib.slabs, (*[intbanksize]int)(ns))
	}

	ib.slabs[slabNo][slabOffset] = offset
}

func (ib *intbank) lookup(sequence uint32) int {
	sequence-- // externally, sequence starts at 1
	slabNo := int(sequence / intbanksize)
	slabOffset := int(sequence % intbanksize)

	return ib.slabs[slabNo][slabOffset]
}

const intbankTag = "INTBANK_V1  "

func (ib *intbank) persist(w io.Writer) error {
	if _, err := w.Write([]byte(intbankTag)); err != nil {
		return fmt.Errorf("writing intbank tag: %w", err)
	}

	data := binary.NativeEndian.AppendUint32(nil, uint32(len(ib.slabs)))
	if _, err := w.Write(data); err != nil {
		return fmt.Errorf("writing intbank slab count: %w", err)
	}

	for _, s := range ib.slabs {
		data := unsafe.Slice((*byte)(unsafe.Pointer(s)), intbanksize*unsafe.Sizeof(int(0)))
		if _, err := w.Write(data); err != nil {
			return fmt.Errorf("writing intbank slab: %w", err)
		}
	}

	return nil
}

func (ib *intbank) load(r io.Reader) error {
	header := make([]byte, len(intbankTag))
	if _, err := io.ReadFull(r, header); err != nil {
		return fmt.Errorf("reading intbank tag: %w", err)
	}
	if string(header) != intbankTag {
		return fmt.Errorf("invalid intbank tag: got %s, want %s", string(header), intbankTag)
	}

	var slabCount uint32
	if err := binary.Read(r, binary.NativeEndian, &slabCount); err != nil {
		return fmt.Errorf("reading intbank slab count: %w", err)
	}

	ib.slabs = make([]*[intbanksize]int, 0, slabCount)
	for range slabCount {
		ns, _ := mmap.Alloc[int](intbanksize)
		slab := (*[intbanksize]int)(ns)
		data := unsafe.Slice((*byte)(unsafe.Pointer(slab)), intbanksize*unsafe.Sizeof(int(0)))
		if _, err := io.ReadFull(r, data); err != nil {
			return fmt.Errorf("reading intbank slab: %w", err)
		}
		ib.slabs = append(ib.slabs, slab)
	}
	return nil
}
