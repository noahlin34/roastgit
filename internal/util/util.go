package util

import (
	"hash/fnv"
	"strings"
	"unicode"
)

// ClampInt clamps v between min and max.
func ClampInt(v, min, max int) int {
	if v < min {
		return min
	}
	if v > max {
		return max
	}
	return v
}

// ShortSHA returns a 7-char short SHA.
func ShortSHA(sha string) string {
	if len(sha) <= 7 {
		return sha
	}
	return sha[:7]
}

// DeterministicChoice picks an option based on a stable hash of seed.
func DeterministicChoice(seed string, options []string) string {
	if len(options) == 0 {
		return ""
	}
	h := fnv.New32a()
	_, _ = h.Write([]byte(seed))
	idx := int(h.Sum32()) % len(options)
	return options[idx]
}

// ContainsAnyWord checks if any word in list is present in s (case-insensitive).
func ContainsAnyWord(s string, words []string) bool {
	lower := strings.ToLower(s)
	for _, w := range words {
		if strings.Contains(lower, w) {
			return true
		}
	}
	return false
}

// IsEmojiOnly returns true if the string has no letters/digits and contains at least one non-ASCII rune.
func IsEmojiOnly(s string) bool {
	trim := strings.TrimSpace(s)
	if trim == "" {
		return false
	}
	hasNonASCII := false
	for _, r := range trim {
		if unicode.IsLetter(r) || unicode.IsDigit(r) {
			return false
		}
		if r > unicode.MaxASCII {
			hasNonASCII = true
		}
	}
	return hasNonASCII
}
