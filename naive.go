package symboltab

// Naive implementation of the same function. Really just intended to compare against
type Naive struct {
	m map[string]int32
	i []string
}

// NewNaive creates a new, basic implementation of the symboltable function
func NewNaive(cap int) *Naive {
	return &Naive{
		m: make(map[string]int32, cap),
		i: make([]string, 0, cap),
	}
}

// StringToSequence converts a string to a sequence number
func (n *Naive) StringToSequence(val string, addNew bool) (seq int32, found bool) {
	seq, ok := n.m[val]
	if ok {
		return seq, true
	}
	if addNew {
		seq := int32(len(n.m)) + 1
		n.i = append(n.i, val)
		n.m[val] = seq
		return seq, false
	}
	return 0, false
}

// SequenceToString retrieves the string for a sequence number
func (n *Naive) SequenceToString(seq int32) string {
	return n.i[seq-1]
}
