package analyze

import (
	"strings"

	"roastgit/internal/util"
)

type MessageInfo struct {
	Generic    bool
	EmojiOnly  bool
	TooLong    bool
	TooShort   bool
	GenericKey string
	Score      int
}

var genericWords = []string{
	"update", "updates", "fix", "fixed", "wip", "tmp", "temp", "stuff",
	"changes", "final", "asdf", "misc", "cleanup", "work", "progress",
	"test", "testing", "oops", "save", "minor", "quick", "hack", "tweak",
}

var genericSet = func() map[string]struct{} {
	m := map[string]struct{}{}
	for _, w := range genericWords {
		m[w] = struct{}{}
	}
	return m
}()

// AnalyzeMessage inspects a subject line for quality signals.
func AnalyzeMessage(subject string) MessageInfo {
	trim := strings.TrimSpace(subject)
	lower := strings.ToLower(trim)
	words := strings.Fields(lower)
	info := MessageInfo{}
	info.EmojiOnly = util.IsEmojiOnly(trim)
	info.TooLong = len([]rune(trim)) > 72
	info.TooShort = len([]rune(trim)) <= 4
	if len(words) > 0 {
		if _, ok := genericSet[words[0]]; ok && len([]rune(trim)) <= 20 {
			info.Generic = true
			info.GenericKey = words[0]
		}
		if _, ok := genericSet[lower]; ok {
			info.Generic = true
			info.GenericKey = lower
		}
	}
	info.Score = scoreMessage(info)
	return info
}

func scoreMessage(info MessageInfo) int {
	score := 100
	if info.Generic {
		score -= 40
	}
	if info.TooShort {
		score -= 20
	}
	if info.TooLong {
		score -= 10
	}
	if info.EmojiOnly {
		score -= 30
	}
	return util.ClampInt(score, 0, 100)
}

// IsLyingMessage checks if a message claims "minor" changes but size is large.
func IsLyingMessage(subject string, linesChanged int, filesChanged int) bool {
	if linesChanged <= 0 && filesChanged <= 0 {
		return false
	}
	lower := strings.ToLower(subject)
	if util.ContainsAnyWord(lower, []string{"minor", "small", "tiny", "quick", "little"}) {
		return linesChanged >= 400 || filesChanged >= 10
	}
	return false
}
