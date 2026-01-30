package analyze

import "testing"

func TestAnalyzeMessageGeneric(t *testing.T) {
	info := AnalyzeMessage("fix")
	if !info.Generic {
		t.Fatalf("expected generic message")
	}
	if !info.TooShort {
		t.Fatalf("expected too short")
	}
}

func TestAnalyzeMessageEmojiOnly(t *testing.T) {
	info := AnalyzeMessage("\U0001F525\U0001F525")
	if !info.EmojiOnly {
		t.Fatalf("expected emoji-only message")
	}
}

func TestAnalyzeMessageTooLong(t *testing.T) {
	msg := "this is a very long commit message that definitely exceeds seventy two characters"
	info := AnalyzeMessage(msg)
	if !info.TooLong {
		t.Fatalf("expected too long message")
	}
}

func TestIsLyingMessage(t *testing.T) {
	if !IsLyingMessage("minor tweak", 500, 12) {
		t.Fatalf("expected lying message")
	}
	if IsLyingMessage("minor tweak", 10, 1) {
		t.Fatalf("expected not lying")
	}
}
