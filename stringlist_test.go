package symboltab

import (
	"os"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestStringList(t *testing.T) {

	sl, err := openStringList("testlist")
	assert.NoError(t, err)

	defer func() {
		os.Remove("testlist.sl")
	}()

	ho, err := sl.Append([]byte("hello"))
	assert.NoError(t, err)

	bo, err := sl.Append([]byte("bye"))
	assert.NoError(t, err)

	hb := sl.Get(ho, 5)
	assert.Equal(t, "hello", string(hb))

	bb := sl.Get(bo, 3)
	assert.Equal(t, "bye", string(bb))

}
