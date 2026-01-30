package roast

import (
	"regexp"
	"strings"
)

var profanityRE = regexp.MustCompile(`(?i)\b(fuck|shit|damn|hell|crap|bastard|ass)\b`)

// Censor replaces profanity with masked versions.
func Censor(s string) string {
	return profanityRE.ReplaceAllStringFunc(s, func(match string) string {
		if match == "" {
			return match
		}
		lower := strings.ToLower(match)
		first := lower[:1]
		return first + "***"
	})
}
