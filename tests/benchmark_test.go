package terlik_test

import (
	"fmt"
	"strings"
	"testing"

	terlik "github.com/KilimcininKorOglu/terlik.go"
)

// spamText builds a ~10KB input packed with repeated short profanity hits —
// the adversarial shape that stresses per-match rescan and mapping costs.
func spamText(unit string, repeats int) string {
	return strings.TrimSpace(strings.Repeat(unit, repeats))
}

func BenchmarkPatternSpam(b *testing.B) {
	tr := mustNewB(b, nil)
	text := spamText("sik test ", 1000) // ~9KB, ~1000 matches
	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	for b.Loop() {
		tr.GetMatches(text, nil)
	}
}

func BenchmarkPatternMixed(b *testing.B) {
	tr := mustNewB(b, nil)
	var sb strings.Builder
	for i := range 500 {
		fmt.Fprintf(&sb, "Bugün hava çok güzel ve s2k kadar soğuk değil, piknik yaptık. Orospu değiliz ama salak bir film izledik %d.", i)
	}
	text := sb.String()
	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	for b.Loop() {
		tr.GetMatches(text, nil)
	}
}

func BenchmarkClean(b *testing.B) {
	tr := mustNewB(b, nil)
	text := spamText("sik test ", 1000)
	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	for b.Loop() {
		tr.Clean(text, nil)
	}
}

func BenchmarkFuzzyLevenshtein(b *testing.B) {
	tr := mustNewB(b, &terlik.Options{
		Language:       "tr",
		EnableFuzzy:    true,
		FuzzyThreshold: 0.8,
	})
	text := spamText("siktiniz salakliklar ustunde calisiyor ", 280) // ~10KB real-ish words
	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	for b.Loop() {
		tr.GetMatches(text, &terlik.DetectOptions{Mode: terlik.ModeLoose})
	}
}

func BenchmarkFuzzyDice(b *testing.B) {
	tr := mustNewB(b, &terlik.Options{
		Language:       "tr",
		EnableFuzzy:    true,
		FuzzyThreshold: 0.8,
		FuzzyAlgorithm: terlik.FuzzyDice,
	})
	text := spamText("siktiniz salakliklar ustunde calisiyor ", 280)
	b.SetBytes(int64(len(text)))
	b.ResetTimer()
	for b.Loop() {
		tr.GetMatches(text, &terlik.DetectOptions{Mode: terlik.ModeLoose})
	}
}

func mustNewB(b *testing.B, opts *terlik.Options) *terlik.Terlik {
	instance, err := terlik.New(opts)
	if err != nil {
		b.Fatalf("New: %v", err)
	}
	return instance
}
