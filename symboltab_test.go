package symboltab

import (
	"strconv"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSymbolTableReverse(t *testing.T) {
	st, err := OpenSymbolTable("test")
	assert.NoError(t, err)
	defer DeleteSymbolTable("test")

	hi, err := st.Add("hat")
	assert.NoError(t, err)

	bi, err := st.Add("bicycle")
	assert.NoError(t, err)

	assert.Equal(t, "hat", st.Reverse(hi))
	assert.Equal(t, "bicycle", st.Reverse(bi))
	assert.Equal(t, "hat", st.Reverse(hi))

}

func BenchmarkAddSymbol(b *testing.B) {

	st, err := OpenSymbolTable("test")
	assert.NoError(b, err)
	defer DeleteSymbolTable("test")

	b.ReportAllocs()
	b.ResetTimer()

	for i := 0; i < b.N; i++ {
		_, err := st.Add(strconv.Itoa(i))
		assert.NoError(b, err)
	}
}

func BenchmarkReverseSymbol_times1000(b *testing.B) {

	st, err := OpenSymbolTable("test")
	assert.NoError(b, err)
	defer DeleteSymbolTable("test")

	for i := 0; i < b.N; i++ {
		_, err := st.Add(strconv.Itoa(i))
		assert.NoError(b, err)
	}

	b.ReportAllocs()
	b.ResetTimer()

	for j := 0; j < 1000; j++ {
		for i := 0; i < b.N; i++ {
			_ = st.Reverse(int32(i))
		}
	}
}
