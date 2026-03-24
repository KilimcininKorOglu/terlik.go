package main

import (
	"math"
	"strings"
	"unicode"
)

// transformFn applies a text transform to a word.
type transformFn func(word string, data *dataSet, lang *langConfig, rand func() float64) string

// transform represents a named transform with its family.
type transform struct {
	fn     transformFn
	name   string
	family string
}

var transforms = []transform{
	{transformSuffix, "suffix", "morphological"},
	{transformCharRepeat, "charRepeat", "repetition"},
	{transformLeet, "leet", "substitution"},
	{transformUnicode, "unicode", "substitution"},
	{transformSeparator, "separator", "separator"},
	{transformSplit, "split", "separator"},
	{transformCase, "case", "casing"},
	{transformZalgo, "zalgo", "obfuscation"},
	{transformZwc, "zwc", "obfuscation"},
	{transformEmojiMix, "emojiMix", "substitution"},
	{transformVowelDrop, "vowelDrop", "morphological"},
	{transformReverse, "reverse", "morphological"},
	{transformDoubling, "doubling", "repetition"},
}

func transformSuffix(word string, data *dataSet, _ *langConfig, rand func() float64) string {
	if len(data.suffixes) == 0 {
		return word
	}
	suffix := data.suffixes[int(math.Floor(rand()*float64(len(data.suffixes))))]
	return word + suffix
}

func transformCharRepeat(word string, _ *dataSet, _ *langConfig, rand func() float64) string {
	runes := []rune(word)
	if len(runes) < 2 {
		return word
	}
	idx := int(math.Floor(rand() * float64(len(runes))))
	times := 2 + int(math.Floor(rand()*3))
	return string(runes[:idx]) + strings.Repeat(string(runes[idx]), times) + string(runes[idx+1:])
}

func transformLeet(word string, data *dataSet, _ *langConfig, rand func() float64) string {
	if len(data.leetMap) == 0 {
		return word
	}
	var b strings.Builder
	for _, ch := range word {
		lower := strings.ToLower(string(ch))
		if opts, ok := data.leetMap[lower]; ok && rand() < 0.4 {
			b.WriteString(opts[int(math.Floor(rand()*float64(len(opts))))])
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func transformUnicode(word string, data *dataSet, _ *langConfig, rand func() float64) string {
	if len(data.unicodeMap) == 0 {
		return word
	}
	var b strings.Builder
	for _, ch := range word {
		lower := strings.ToLower(string(ch))
		if opts, ok := data.unicodeMap[lower]; ok && rand() < 0.35 {
			b.WriteString(opts[int(math.Floor(rand()*float64(len(opts))))])
		} else {
			b.WriteRune(ch)
		}
	}
	return b.String()
}

func transformSeparator(word string, data *dataSet, _ *langConfig, rand func() float64) string {
	runes := []rune(word)
	if len(data.separators) == 0 || len(runes) < 2 {
		return word
	}
	sep := data.separators[int(math.Floor(rand()*float64(len(data.separators))))]
	parts := make([]string, len(runes))
	for i, r := range runes {
		parts[i] = string(r)
	}
	return strings.Join(parts, sep)
}

func transformSplit(word string, _ *dataSet, _ *langConfig, rand func() float64) string {
	runes := []rune(word)
	if len(runes) < 3 {
		return word
	}
	pos := 1 + int(math.Floor(rand()*float64(len(runes)-1)))
	return string(runes[:pos]) + " " + string(runes[pos:])
}

func transformCase(word string, _ *dataSet, _ *langConfig, rand func() float64) string {
	mode := rand()
	runes := []rune(word)
	if mode < 0.33 {
		return strings.ToUpper(word)
	}
	if mode < 0.66 {
		var b strings.Builder
		for _, r := range runes {
			if rand() < 0.5 {
				b.WriteRune(unicode.ToUpper(r))
			} else {
				b.WriteRune(unicode.ToLower(r))
			}
		}
		return b.String()
	}
	// alternating
	var b strings.Builder
	for i, r := range runes {
		if i%2 == 0 {
			b.WriteRune(unicode.ToUpper(r))
		} else {
			b.WriteRune(unicode.ToLower(r))
		}
	}
	return b.String()
}

func transformZalgo(word string, data *dataSet, _ *langConfig, rand func() float64) string {
	if len(data.zalgoChars) == 0 {
		return word
	}
	var b strings.Builder
	for _, ch := range word {
		b.WriteRune(ch)
		count := 1 + int(math.Floor(rand()*3))
		for j := 0; j < count; j++ {
			b.WriteString(data.zalgoChars[int(math.Floor(rand()*float64(len(data.zalgoChars))))])
		}
	}
	return b.String()
}

func transformZwc(word string, data *dataSet, _ *langConfig, rand func() float64) string {
	runes := []rune(word)
	if len(data.zwcChars) == 0 || len(runes) < 2 {
		return word
	}
	var b strings.Builder
	for i, r := range runes {
		b.WriteRune(r)
		if i < len(runes)-1 && rand() < 0.5 {
			b.WriteString(data.zwcChars[int(math.Floor(rand()*float64(len(data.zwcChars))))])
		}
	}
	return b.String()
}

func transformEmojiMix(word string, data *dataSet, _ *langConfig, rand func() float64) string {
	runes := []rune(word)
	if len(data.emojiReplacements) == 0 || len(runes) < 2 {
		return word
	}
	emoji := data.emojiReplacements[int(math.Floor(rand()*float64(len(data.emojiReplacements))))]
	pos := 1 + int(math.Floor(rand()*float64(len(runes)-1)))
	return string(runes[:pos]) + emoji + string(runes[pos:])
}

func transformVowelDrop(word string, _ *dataSet, lang *langConfig, rand func() float64) string {
	runes := []rune(word)
	if len(runes) < 3 {
		return word
	}
	var b strings.Builder
	dropped := false
	for i, r := range runes {
		if strings.ContainsRune(lang.vowels, unicode.ToLower(r)) && i > 0 && i < len(runes)-1 && rand() < 0.5 {
			dropped = true
			continue
		}
		b.WriteRune(r)
	}
	if dropped {
		return b.String()
	}
	return word
}

func transformReverse(word string, _ *dataSet, _ *langConfig, _ func() float64) string {
	runes := []rune(word)
	for i, j := 0, len(runes)-1; i < j; i, j = i+1, j-1 {
		runes[i], runes[j] = runes[j], runes[i]
	}
	return string(runes)
}

func transformDoubling(word string, _ *dataSet, _ *langConfig, rand func() float64) string {
	runes := []rune(word)
	if len(runes) < 2 {
		return word
	}
	idx := int(math.Floor(rand() * float64(len(runes))))
	return string(runes[:idx+1]) + string(runes[idx]) + string(runes[idx+1:])
}
