package util

import "testing"

func TestSampleIndicesDeterministic(t *testing.T) {
	idx1 := SampleIndices(100, 10)
	idx2 := SampleIndices(100, 10)
	if len(idx1) != len(idx2) {
		t.Fatalf("expected same length")
	}
	for i := range idx1 {
		if idx1[i] != idx2[i] {
			t.Fatalf("expected deterministic sampling")
		}
	}
	if idx1[0] != 0 || idx1[len(idx1)-1] != 99 {
		t.Fatalf("expected endpoints sampled")
	}
}

func TestSampleIndicesAll(t *testing.T) {
	idx := SampleIndices(5, 10)
	if len(idx) != 5 {
		t.Fatalf("expected all indices")
	}
	for i := range idx {
		if idx[i] != i {
			t.Fatalf("expected index %d, got %d", i, idx[i])
		}
	}
}
