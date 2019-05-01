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

	assertStringToSequence := func(seq int32, existing bool, val string) {
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

	for i := 0; i < 10000; i++ {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		assert.False(t, found)
		assert.Equal(t, int32(i+1), seq)
	}

	for i := 0; i < 10000; i++ {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		assert.True(t, found)
		assert.Equal(t, int32(i+1), seq)
	}

	for i := 0; i < 10000; i++ {
		str := st.SequenceToString(int32(i + 1))
		assert.Equal(t, strconv.Itoa(i), str)
	}
}

func TestGrowth2(t *testing.T) {
	st := New(16)

	for i := 0; i < 10000; i++ {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		assert.False(t, found)
		assert.Equal(t, int32(i+1), seq)

		seq, found = st.StringToSequence(strconv.Itoa(i), true)
		assert.True(t, found)
		assert.Equal(t, int32(i+1), seq)
	}
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

func BenchmarkSequenceToString(b *testing.B) {
	st := New(b.N)
	for i := 0; i < b.N; i++ {
		st.StringToSequence(strconv.Itoa(i), true)
	}

	b.ReportAllocs()
	b.ResetTimer()

	var str string
	for i := 1; i <= b.N; i++ {
		str = st.SequenceToString(int32(i))
	}

	if str != strconv.Itoa(b.N-1) {
		b.Errorf("last symbol doesn't match - get %s", str)
	}
}

func BenchmarkExisting(b *testing.B) {
	st := New(b.N)
	values := make([]string, b.N)
	for i := range values {
		values[i] = strconv.Itoa(i)
	}

	for _, val := range values {
		st.StringToSequence(val, true)
	}

	b.ReportAllocs()
	b.ResetTimer()

	var seq int32
	for _, val := range values {
		seq, _ = st.StringToSequence(val, false)
	}

	if st.SequenceToString(seq) != strconv.Itoa(b.N-1) {
		b.Errorf("last symbol doesn't match - get %s", st.SequenceToString(seq))
	}
}

func BenchmarkMiss(b *testing.B) {
	st := New(b.N)
	values := make([]string, b.N)
	for i := range values {
		values[i] = strconv.Itoa(i)
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
	seq, found := st.StringToSequence("10293-ahdb-28383-555", true)
	fmt.Println(found)
	fmt.Println(st.SequenceToString(seq))
	// Output: false
	// 10293-ahdb-28383-555
}

func BenchmarkMakeBigSlice(b *testing.B) {
	for i := 0; i < b.N; i++ {
		sl := make([]int32, 1e8)
		runtime.KeepAlive(sl)
	}
}
