package terlik

import (
	"strings"
	"unicode"
	"unicode/utf8"

	"golang.org/x/text/unicode/norm"
)

// invisibleChars contains zero-width and invisible Unicode characters used to bypass detection.
var invisibleChars = map[rune]bool{
	'\u200B': true, // Zero-width space
	'\u200C': true, // ZWNJ
	'\u200D': true, // ZWJ
	'\u200E': true, // LRM
	'\u200F': true, // RLM
	'\uFEFF': true, // BOM
	'\u00AD': true, // Soft hyphen
	'\u034F': true, // Combining grapheme joiner
	'\u2060': true, // Word joiner
	'\u2061': true, // Function application
	'\u2062': true, // Invisible times
	'\u2063': true, // Invisible separator
	'\u2064': true, // Invisible plus
	'\u180E': true, // Mongolian vowel separator
}

// cyrillicConfusables maps visually identical Cyrillic chars to Latin equivalents.
var cyrillicConfusables = map[rune]rune{
	'\u0430': 'a', // Cyrillic а → Latin a
	'\u0441': 'c', // Cyrillic с → Latin c
	'\u0435': 'e', // Cyrillic е → Latin e
	'\u0456': 'i', // Cyrillic і → Latin i
	'\u043E': 'o', // Cyrillic о → Latin o
	'\u0440': 'p', // Cyrillic р → Latin p
	'\u0443': 'u', // Cyrillic у → Latin u
	'\u0445': 'x', // Cyrillic х → Latin x
}

// punctuationChars contains punctuation that may appear between letters.
var punctuationChars = map[rune]bool{
	'.': true, '-': true, '_': true, '*': true,
	',': true, ';': true, ':': true, '!': true, '?': true,
}

// isExtendedLetter checks if a rune matches [a-zA-ZÀ-ɏ].
func isExtendedLetter(r rune) bool {
	return (r >= 'a' && r <= 'z') || (r >= 'A' && r <= 'Z') || (r >= 0x00C0 && r <= 0x024F)
}

// isWordChar checks if a rune matches [a-zA-Z0-9À-ɏ].
func isWordChar(r rune) bool {
	return isExtendedLetter(r) || (r >= '0' && r <= '9')
}

// localeLowerCase performs locale-aware lowercasing.
func localeLowerCase(s string, locale string) string {
	switch locale {
	case "tr", "az":
		return turkishToLower(s)
	default:
		return strings.ToLower(s)
	}
}

// turkishToLower handles Turkish-specific case folding: İ→i, I→ı.
func turkishToLower(s string) string {
	var b strings.Builder
	b.Grow(len(s) + utf8.UTFMax)
	for _, r := range s {
		switch r {
		case 'I':
			b.WriteRune('\u0131') // ı
		case '\u0130': // İ
			b.WriteRune('i')
		default:
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

// stripInvisibleChars removes zero-width and invisible characters (stage 1).
func stripInvisibleChars(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if !invisibleChars[r] {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// nfkdDecompose applies NFKD Unicode decomposition (stage 2).
func nfkdDecompose(text string) string {
	return norm.NFKD.String(text)
}

// stripCombiningMarks removes combining diacritical marks U+0300-U+036F (stage 3).
func stripCombiningMarks(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if r < 0x0300 || r > 0x036F {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// replaceFromMap replaces characters using a lookup map (stages 5, 6, 8).
func replaceFromMap(text string, m map[string]string) string {
	if len(m) == 0 {
		return text
	}
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		ch := string(r)
		if replacement, ok := m[ch]; ok {
			b.WriteString(replacement)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// replaceCyrillicConfusables maps Cyrillic lookalikes to Latin (stage 5).
func replaceCyrillicConfusables(text string) string {
	var b strings.Builder
	b.Grow(len(text))
	for _, r := range text {
		if latin, ok := cyrillicConfusables[r]; ok {
			b.WriteRune(latin)
		} else {
			b.WriteRune(r)
		}
	}
	return b.String()
}

// expandNumbers replaces number strings between letters with their word forms (stage 7).
// Uses manual scanning since Go RE2 doesn't support lookbehind.
func expandNumbers(text string, expansions [][2]string) string {
	if len(expansions) == 0 {
		return text
	}

	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return text
	}

	var b strings.Builder
	b.Grow(len(text) + 32)
	i := 0
	for i < n {
		if i > 0 && isExtendedLetter(runes[i-1]) {
			// Try each expansion (longest first — expansions are already ordered by length desc)
			if word, width, ok := tryExpandAt(runes, i, expansions); ok {
				b.WriteString(word)
				i += width
				continue
			}
		}
		b.WriteRune(runes[i])
		i++
	}

	return b.String()
}

// tryExpandAt finds the first expansion whose number string matches at position i
// and is followed by a letter, returning its word form and rune width.
func tryExpandAt(runes []rune, i int, expansions [][2]string) (string, int, bool) {
	for _, exp := range expansions {
		numRunes := []rune(exp[0])
		if !matchNumberAt(runes, i, numRunes) {
			continue
		}
		// Check if followed by a letter
		if i+len(numRunes) < len(runes) && isExtendedLetter(runes[i+len(numRunes)]) {
			return exp[1], len(numRunes), true
		}
	}
	return "", 0, false
}

// matchNumberAt reports whether num matches runes starting at position i.
func matchNumberAt(runes []rune, i int, num []rune) bool {
	if i+len(num) > len(runes) {
		return false
	}
	for k, r := range num {
		if runes[i+k] != r {
			return false
		}
	}
	return true
}

// removePunctuationBetweenLetters strips punctuation when surrounded by letters (stage 9).
// Manual implementation since Go RE2 doesn't support lookbehind.
func removePunctuationBetweenLetters(text string) string {
	runes := []rune(text)
	n := len(runes)
	if n == 0 {
		return text
	}

	var b strings.Builder
	b.Grow(len(text))
	i := 0
	for i < n {
		if !punctuationChars[runes[i]] {
			b.WriteRune(runes[i])
			i++
			continue
		}
		start := i
		i = punctuationRun(runes, i)
		if isPunctuationBetweenLetters(runes, start, i) {
			continue // Drop the punctuation run
		}
		for k := start; k < i; k++ {
			b.WriteRune(runes[k])
		}
	}

	return b.String()
}

// punctuationRun returns the end index (exclusive) of the punctuation run starting at i.
func punctuationRun(runes []rune, i int) int {
	for i < len(runes) && punctuationChars[runes[i]] {
		i++
	}
	return i
}

// isPunctuationBetweenLetters reports whether the run runes[start:end] sits between
// two extended letters and should be dropped.
func isPunctuationBetweenLetters(runes []rune, start, end int) bool {
	return start > 0 && isExtendedLetter(runes[start-1]) && end < len(runes) && isExtendedLetter(runes[end])
}

// collapseRepeats reduces 3+ consecutive identical characters to 1 (stage 10).
// Manual implementation since Go RE2 doesn't support backreferences (.)\1{2,}.
func collapseRepeats(text string) string {
	runes := []rune(text)
	n := len(runes)
	if n <= 2 {
		return text
	}

	var b strings.Builder
	b.Grow(len(text))
	i := 0

	for i < n {
		ch := runes[i]
		// Count consecutive identical characters
		j := i + 1
		for j < n && runes[j] == ch {
			j++
		}
		count := j - i

		if count >= 3 {
			// Collapse to single character
			b.WriteRune(ch)
		} else {
			// Keep all (1 or 2)
			for k := i; k < j; k++ {
				b.WriteRune(ch)
			}
		}
		i = j
	}

	return b.String()
}

// trimWhitespace normalizes whitespace runs to single space and trims (stage 10).
func trimWhitespace(text string) string {
	return strings.TrimSpace(strings.Join(strings.Fields(text), " "))
}

// CreateNormalizer creates a language-specific normalize function using the given config.
// Returns a function that applies a 10-stage normalization pipeline.
func CreateNormalizer(config NormalizerConfig) func(string) string {
	return func(text string) string {
		result := text
		// Stage 1: Strip invisible chars
		result = stripInvisibleChars(result)
		// Stage 2: NFKD decompose
		result = nfkdDecompose(result)
		// Stage 3: Strip combining marks
		result = stripCombiningMarks(result)
		// Stage 4: Locale-aware lowercase
		result = localeLowerCase(result, config.Locale)
		// Stage 5: Cyrillic confusables
		result = replaceCyrillicConfusables(result)
		// Stage 6: Language-specific char folding
		result = replaceFromMap(result, config.CharMap)
		// Stage 7: Number expansion
		result = expandNumbers(result, config.NumberExpansions)
		// Stage 8: Leet decode
		result = replaceFromMap(result, config.LeetMap)
		// Stage 9: Punctuation removal between letters
		result = removePunctuationBetweenLetters(result)
		// Stage 10: Repeat collapse + whitespace trim
		result = collapseRepeats(result)
		result = trimWhitespace(result)
		return result
	}
}

// Normalize applies the default Turkish locale normalization pipeline.
func Normalize(text string) string {
	return defaultTurkishNormalize(text)
}

var defaultTurkishNormalize = CreateNormalizer(NormalizerConfig{
	Locale: "tr",
	CharMap: map[string]string{
		"ç": "c", "Ç": "c", "ğ": "g", "Ğ": "g", "ı": "i", "İ": "i",
		"ö": "o", "Ö": "o", "ş": "s", "Ş": "s", "ü": "u", "Ü": "u",
	},
	LeetMap: map[string]string{
		"0": "o", "1": "i", "2": "i", "3": "e", "4": "a",
		"5": "s", "6": "g", "7": "t", "8": "b", "9": "g",
		"@": "a", "$": "s", "!": "i",
	},
	NumberExpansions: [][2]string{
		{"100", "yuz"}, {"50", "elli"}, {"10", "on"}, {"2", "iki"},
	},
})
