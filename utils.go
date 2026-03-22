package terlik

import "unicode/utf8"

// MaxInputLength is the default maximum input length (10,000 characters).
const MaxInputLength = 10_000

// validateInput sanitizes text input with length truncation.
func validateInput(text string, maxLength int) string {
	if len(text) == 0 {
		return ""
	}
	runeCount := utf8.RuneCountInString(text)
	if runeCount > maxLength {
		runes := []rune(text)
		return string(runes[:maxLength])
	}
	return text
}
