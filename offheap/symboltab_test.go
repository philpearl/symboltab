package offheap

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestBasic(t *testing.T) {
	st := New(16)
	defer st.Close()

	assertStringToSequence := func(seq uint32, existing bool, val string) {
		t.Helper()
		seqa, existinga := st.StringToSequence(val, true)
		assert.Equal(t, existing, existinga)
		if existinga {
			assert.Equal(t, seq, seqa)
		}
	}

	assert.Zero(t, st.SymbolSize())

	assertStringToSequence(1, false, "a1")
	assertStringToSequence(2, false, "a2")
	assertStringToSequence(3, false, "a3")
	assertStringToSequence(2, true, "a2")
	assertStringToSequence(3, true, "a3")

	assert.Equal(t, 1<<18, st.SymbolSize())

	assert.Equal(t, "a1", st.SequenceToString(1))
	assert.Equal(t, "a2", st.SequenceToString(2))
	assert.Equal(t, "a3", st.SequenceToString(3))
}

func TestGrowth(t *testing.T) {
	st := New(16)
	defer st.Close()

	for i := range 10000 {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		assert.False(t, found)
		assert.Equal(t, uint32(i+1), seq)
	}

	for i := range 10000 {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		assert.True(t, found)
		assert.Equal(t, uint32(i+1), seq)
	}

	for i := range 10000 {
		str := st.SequenceToString(uint32(i + 1))
		assert.Equal(t, strconv.Itoa(i), str)
	}
}

func TestGrowth2(t *testing.T) {
	st := New(16)
	defer st.Close()

	for i := range 10000 {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		assert.False(t, found)
		assert.Equal(t, uint32(i+1), seq)

		seq, found = st.StringToSequence(strconv.Itoa(i), true)
		assert.True(t, found)
		assert.Equal(t, uint32(i+1), seq)
	}
}

func TestAddNew(t *testing.T) {
	st := New(16)
	defer st.Close()
	// Won't add entry if asked not to
	seq, existing := st.StringToSequence("hat", false)
	assert.False(t, existing)
	assert.Equal(t, uint32(0), seq)

	seq, existing = st.StringToSequence("hat", true)
	assert.False(t, existing)
	assert.Equal(t, uint32(1), seq)

	// Can find existing entry if not asked to add new
	seq, existing = st.StringToSequence("hat", false)
	assert.True(t, existing)
	assert.Equal(t, uint32(1), seq)
}

func TestLowGC(t *testing.T) {
	st := New(16)
	defer st.Close()
	for i := range 10000000 {
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
	st := New(16)
	defer st.Close()
	for _, sym := range symbols {
		st.StringToSequence(sym, true)
	}

	if symbols[0] != st.SequenceToString(1) {
		b.Errorf("first symbol doesn't match - get %s", st.SequenceToString(1))
	}
}

func BenchmarkSymbolTabSmall(b *testing.B) {
	for _, len := range []int{10_000, 100_000, 1_000_000, 10_000_000, 100_000_000} {
		b.Run(strconv.Itoa(len), func(b *testing.B) {
			symbols := make([]string, len)
			for i := range symbols {
				symbols[i] = strconv.Itoa(i)
			}

			b.ReportAllocs()
			b.ResetTimer()

			for b.Loop() {
				st := New(16)
				for _, sym := range symbols {
					st.StringToSequence(sym, true)
				}
				st.Close()
			}
		})
	}
}

func BenchmarkSequenceToString(b *testing.B) {
	st := New(16)
	defer st.Close()
	for i := range 100_000 {
		st.StringToSequence(strconv.Itoa(i), true)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for b.Loop() {
		var str string
		for i := range 100_000 {
			str = st.SequenceToString(uint32(i + 1))
		}

		if str != strconv.Itoa(100_000-1) {
			b.Errorf("last symbol doesn't match - get %s", str)
		}
	}
}

func BenchmarkExisting(b *testing.B) {
	st := New(b.N)
	defer st.Close()
	values := make([]string, b.N)
	for i := range values {
		values[i] = strconv.Itoa(i)
	}

	for _, val := range values {
		st.StringToSequence(val, true)
	}

	b.ReportAllocs()
	b.ResetTimer()

	var seq uint32
	for _, val := range values {
		seq, _ = st.StringToSequence(val, false)
	}

	if st.SequenceToString(seq) != strconv.Itoa(b.N-1) {
		b.Errorf("last symbol doesn't match - get %s", st.SequenceToString(seq))
	}
}

func BenchmarkMiss(b *testing.B) {
	st := New(b.N)
	defer st.Close()

	// We want some entries in the table to make misses a bit more realistic.
	for i := range 10_000 {
		st.StringToSequence(strconv.Itoa(i), true)
	}

	values := make([]string, b.N)
	for i := range values {
		values[i] = strconv.Itoa(i + 10_000)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for _, val := range values {
		_, found := st.StringToSequence(val, false)
		if found {
			b.Errorf("found value %s", val)
		}
	}
}

func ExampleSymbolTab() {
	st := SymbolTab{}
	defer st.Close()
	seq, found := st.StringToSequence("10293-ahdb-28383-555", true)
	fmt.Println(found)
	fmt.Println(st.SequenceToString(seq))
	// Output: false
	// 10293-ahdb-28383-555
}

func BenchmarkMakeBigSlice(b *testing.B) {
	b.ReportAllocs()
	for b.Loop() {
		sl := make([]int32, 1e8)
		runtime.KeepAlive(sl)
	}
}
