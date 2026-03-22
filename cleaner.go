package terlik

import (
	"sort"
	"strings"
	"unicode/utf8"
)

func maskStars(word string) string {
	return strings.Repeat("*", utf8.RuneCountInString(word))
}

func maskPartial(word string) string {
	runes := []rune(word)
	n := len(runes)
	if n <= 2 {
		return strings.Repeat("*", n)
	}
	return string(runes[0]) + strings.Repeat("*", n-2) + string(runes[n-1])
}

// ApplyMask applies a mask to a single word using the specified style.
func ApplyMask(word string, style MaskStyle, replaceMask string) string {
	switch style {
	case MaskStars:
		return maskStars(word)
	case MaskPartial:
		return maskPartial(word)
	case MaskReplace:
		return replaceMask
	default:
		return maskStars(word)
	}
}

// CleanText replaces all matched profanity in the text with masked versions.
func CleanText(text string, matches []MatchResult, style MaskStyle, replaceMask string) string {
	if len(matches) == 0 {
		return text
	}

	sorted := make([]MatchResult, len(matches))
	copy(sorted, matches)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Index > sorted[j].Index
	})

	result := text
	for _, m := range sorted {
		masked := ApplyMask(m.Word, style, replaceMask)
		end := m.Index + len(m.Word)
		if m.Index >= 0 && end <= len(result) {
			result = result[:m.Index] + masked + result[end:]
		}
	}

	return result
}
