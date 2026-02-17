package symboltab

import (
	"fmt"
	"runtime"
	"strconv"
	"testing"
	"time"
)

func TestBasic(t *testing.T) {
	st := New(16)

	assertStringToSequence := func(seq uint32, existing bool, val string) {
		t.Helper()
		seqa, existinga := st.StringToSequence(val, true)
		if existinga != existing {
			t.Errorf("expected existing=%v, got %v for %q", existing, existinga, val)
		}
		if existinga && seqa != seq {
			t.Errorf("expected seq=%d, got %d for %q", seq, seqa, val)
		}
	}

	if st.SymbolSize() != 0 {
		t.Errorf("expected SymbolSize=0, got %d", st.SymbolSize())
	}

	assertStringToSequence(1, false, "a1")
	assertStringToSequence(2, false, "a2")
	assertStringToSequence(3, false, "a3")
	assertStringToSequence(2, true, "a2")
	assertStringToSequence(3, true, "a3")

	if st.SymbolSize() != 1<<18 {
		t.Errorf("expected SymbolSize=%d, got %d", 1<<18, st.SymbolSize())
	}

	if got := st.SequenceToString(1); got != "a1" {
		t.Errorf("expected a1, got %s", got)
	}
	if got := st.SequenceToString(2); got != "a2" {
		t.Errorf("expected a2, got %s", got)
	}
	if got := st.SequenceToString(3); got != "a3" {
		t.Errorf("expected a3, got %s", got)
	}
}

func TestInsertString(t *testing.T) {
	st := New(16)

	for i := range 10000 {
		seq := st.InsertString(strconv.Itoa(i))
		if seq != uint32(i+1) {
			t.Errorf("expected sequence %d, got %d", i+1, seq)
		}
	}

	for i := range 10000 {
		str := st.SequenceToString(uint32(i + 1))
		if str != strconv.Itoa(i) {
			t.Errorf("expected string %s, got %s", strconv.Itoa(i), str)
		}
	}
}

func TestGrowth(t *testing.T) {
	st := New(16)

	for i := range 10_000 {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		if found {
			t.Errorf("expected not found for %d", i)
		}
		if seq != uint32(i+1) {
			t.Errorf("expected seq=%d, got %d", i+1, seq)
		}
	}

	for i := range 10_000 {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		if !found {
			t.Errorf("expected found for %d", i)
		}
		if seq != uint32(i+1) {
			t.Errorf("expected seq=%d, got %d", i+1, seq)
		}
	}

	for i := range 10_000 {
		str := st.SequenceToString(uint32(i + 1))
		if str != strconv.Itoa(i) {
			t.Errorf("expected %s, got %s", strconv.Itoa(i), str)
		}
	}
}

func TestGrowth2(t *testing.T) {
	st := New(16)

	for i := range 10_000 {
		seq, found := st.StringToSequence(strconv.Itoa(i), true)
		if found {
			t.Errorf("expected not found for %d", i)
		}
		if seq != uint32(i+1) {
			t.Errorf("expected seq=%d, got %d", i+1, seq)
		}

		seq, found = st.StringToSequence(strconv.Itoa(i), true)
		if !found {
			t.Errorf("expected found for %d", i)
		}
		if seq != uint32(i+1) {
			t.Errorf("expected seq=%d, got %d", i+1, seq)
		}
	}
}

func TestAddNew(t *testing.T) {
	st := New(16)
	// Won't add entry if asked not to
	seq, existing := st.StringToSequence("hat", false)
	if existing {
		t.Error("expected not existing")
	}
	if seq != 0 {
		t.Errorf("expected seq=0, got %d", seq)
	}

	seq, existing = st.StringToSequence("hat", true)
	if existing {
		t.Error("expected not existing")
	}
	if seq != 1 {
		t.Errorf("expected seq=1, got %d", seq)
	}

	// Can find existing entry if not asked to add new
	seq, existing = st.StringToSequence("hat", false)
	if !existing {
		t.Error("expected existing")
	}
	if seq != 1 {
		t.Errorf("expected seq=1, got %d", seq)
	}
}

func TestLowGC(t *testing.T) {
	st := New(16)
	for i := 0; i < 1e7; i++ {
		st.StringToSequence(strconv.Itoa(i), true)
	}
	runtime.GC()
	start := time.Now()
	runtime.GC()
	if elapsed := time.Since(start); elapsed >= time.Millisecond*5 {
		t.Errorf("GC took too long: %v", elapsed)
	}

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
		str = st.SequenceToString(uint32(i))
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
