package git

import (
	"os"
	"strings"
	"testing"
)

func TestParseNumstatFromFixture(t *testing.T) {
	data, err := os.ReadFile("fixtures/numstat_basic.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	text := strings.ReplaceAll(string(data), "<US>", string([]byte{unitSep}))
	stats, err := ParseNumstat(strings.NewReader(text))
	if err != nil {
		t.Fatalf("parse numstat: %v", err)
	}
	if len(stats) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(stats))
	}
	first := stats["1111111111111111111111111111111111111111"]
	if first.Files != 2 || first.BinaryFiles != 1 {
		t.Fatalf("unexpected stats: %+v", first)
	}
}
