package symboltab

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestIntbank(t *testing.T) {
	ib := intbank{}
	ib.save(1, 37)
	ib.save(2, 43)

	assert.EqualValues(t, 37, ib.lookup(1))
	assert.EqualValues(t, 43, ib.lookup(2))
	assert.EqualValues(t, 37, ib.lookup(1))
}
