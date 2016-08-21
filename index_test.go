package symboltab

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIndex(t *testing.T) {

	idx, err := openIndex("testindex")
	assert.NoError(t, err)

	defer func() {
		os.Remove("testindex.idx")
	}()

	assert.NoError(t, idx.Put(0, 5))
	assert.NoError(t, idx.Put(5, 3))
	assert.NoError(t, idx.Put(8, 100))

	assertOffsetLength := func(i int, offset, length int32) {
		o, l := idx.Get(i)
		assert.Equal(t, offset, o)
		assert.Equal(t, length, l)
	}

	assertOffsetLength(1, 5, 3)
	assertOffsetLength(0, 0, 5)
	assertOffsetLength(2, 8, 100)

}
