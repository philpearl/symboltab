package symboltab

import "testing"

func TestIntbank(t *testing.T) {
	ib := intbank{}
	ib.save(1, 37)
	ib.save(2, 43)

	if got := ib.lookup(1); got != 37 {
		t.Errorf("expected 37, got %d", got)
	}
	if got := ib.lookup(2); got != 43 {
		t.Errorf("expected 43, got %d", got)
	}
	if got := ib.lookup(1); got != 37 {
		t.Errorf("expected 37, got %d", got)
	}
}
