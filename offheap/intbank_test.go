package offheap

import (
	"io"
	"os"
	"testing"
)

func TestIntbank(t *testing.T) {
	var ib intbank
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

func TestIntbankReload(t *testing.T) {
	var ib intbank
	for i := range 100_000 {
		ib.save(uint32(i+1), i*10)
	}

	dir := t.TempDir()
	path := dir + "/intbank.dat"
	w, err := os.Create(path)
	if err != nil {
		t.Fatalf("failed to open file: %v", err)
	}
	if err := ib.persist(w); err != nil {
		t.Fatalf("failed to persist intbank: %v", err)
	}
	if _, err := w.Seek(0, io.SeekStart); err != nil {
		t.Fatalf("failed to seek intbank file: %v", err)
	}
	var ib2 intbank
	if err := ib2.load(w); err != nil {
		t.Fatalf("failed to load intbank: %v", err)
	}

	for i := range 100_000 {
		expected := i * 10
		if got := ib2.lookup(uint32(i + 1)); got != expected {
			t.Errorf("expected %d, got %d for sequence %d", expected, got, i+1)
		}
	}
}
