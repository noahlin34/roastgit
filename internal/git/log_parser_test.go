package git

import (
	"os"
	"strings"
	"testing"
)

func TestParseLogFromFixture(t *testing.T) {
	data, err := os.ReadFile("fixtures/log_basic.txt")
	if err != nil {
		t.Fatalf("read fixture: %v", err)
	}
	text := strings.ReplaceAll(string(data), "<US>", string([]byte{unitSep}))
	text = strings.ReplaceAll(text, "<RS>", string([]byte{recordSep}))
	commits, err := ParseLog(strings.NewReader(text))
	if err != nil {
		t.Fatalf("parse log: %v", err)
	}
	if len(commits) != 2 {
		t.Fatalf("expected 2 commits, got %d", len(commits))
	}
	if commits[0].SHA == "" || commits[1].AuthorEmail == "" {
		t.Fatalf("expected fields populated")
	}
}
