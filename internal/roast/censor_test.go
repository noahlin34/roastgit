package roast

import "testing"

func TestCensor(t *testing.T) {
	out := Censor("This is shit and FUCK and damn")
	expected := "This is s*** and f*** and d***"
	if out != expected {
		t.Fatalf("expected %q, got %q", expected, out)
	}
}
