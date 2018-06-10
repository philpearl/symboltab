package symboltab

import (
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	st := New(16)

	assertStringToSequence := func(seq int32, existing bool, val string) {
		t.Helper()
		seqa, existinga := st.StringToSequence(val, true)
		assert.Equal(t, existing, existinga)
		if existinga {
			assert.Equal(t, seq, seqa)
		}
	}

	assertStringToSequence(1, false, "a1")
	assertStringToSequence(2, false, "a2")
	assertStringToSequence(3, false, "a3")
	assertStringToSequence(2, true, "a2")
	assertStringToSequence(3, true, "a3")

	assert.Equal(t, "a1", st.SequenceToString(1))
	assert.Equal(t, "a2", st.SequenceToString(2))
	assert.Equal(t, "a3", st.SequenceToString(3))
}

func TestAddNew(t *testing.T) {
	st := New(16)
	// Won't add entry if asked not to
	seq, existing := st.StringToSequence("hat", false)
	assert.False(t, existing)
	assert.Equal(t, int32(0), seq)

	seq, existing = st.StringToSequence("hat", true)
	assert.False(t, existing)
	assert.Equal(t, int32(1), seq)

	// Can find existing entry if not asked to add new
	seq, existing = st.StringToSequence("hat", false)
	assert.True(t, existing)
	assert.Equal(t, int32(1), seq)
}

func TestLowGC(t *testing.T) {
	st := New(16)
	for i := 0; i < 1E7; i++ {
		st.StringToSequence(strconv.Itoa(i), true)
	}
	runtime.GC()
	start := time.Now()
	runtime.GC()
	assert.True(t, time.Since(start) < time.Millisecond*5)

	runtime.KeepAlive(st)
}

func BenchmarkSymbolTab(b *testing.B) {
	symbols := make([]string, b.N)
	for i := range symbols {
		symbols[i] = strconv.Itoa(i)
	}

	b.ReportAllocs()
	b.ResetTimer()
	st := New(b.N)
	for _, sym := range symbols {
		st.StringToSequence(sym, true)
	}

	if symbols[0] != st.SequenceToString(1) {
		b.Errorf("first symbol doesn't match - get %s", st.SequenceToString(1))
	}
}
