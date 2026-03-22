package terlik_test

import (
	"math"
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func TestLevenshteinDistanceTests(t *testing.T) {
	t.Run("returns 0 for identical strings", func(t *testing.T) {
		if d := terlik.LevenshteinDistance("abc", "abc"); d != 0 {
			t.Errorf("got %d, want 0", d)
		}
	})
	t.Run("returns correct distance for single edit", func(t *testing.T) {
		if d := terlik.LevenshteinDistance("abc", "ab"); d != 1 {
			t.Errorf("abc→ab: got %d, want 1", d)
		}
		if d := terlik.LevenshteinDistance("abc", "axc"); d != 1 {
			t.Errorf("abc→axc: got %d, want 1", d)
		}
		if d := terlik.LevenshteinDistance("abc", "abcd"); d != 1 {
			t.Errorf("abc→abcd: got %d, want 1", d)
		}
	})
	t.Run("handles empty strings", func(t *testing.T) {
		if d := terlik.LevenshteinDistance("", "abc"); d != 3 {
			t.Errorf("→abc: got %d, want 3", d)
		}
		if d := terlik.LevenshteinDistance("abc", ""); d != 3 {
			t.Errorf("abc→: got %d, want 3", d)
		}
		if d := terlik.LevenshteinDistance("", ""); d != 0 {
			t.Errorf("→: got %d, want 0", d)
		}
	})
	t.Run("returns correct distance for multiple edits", func(t *testing.T) {
		if d := terlik.LevenshteinDistance("kitten", "sitting"); d != 3 {
			t.Errorf("kitten→sitting: got %d, want 3", d)
		}
	})
}

func TestLevenshteinSimilarityTests(t *testing.T) {
	t.Run("returns 1 for identical strings", func(t *testing.T) {
		if s := terlik.LevenshteinSimilarity("abc", "abc"); s != 1.0 {
			t.Errorf("got %f, want 1.0", s)
		}
	})
	t.Run("returns ~0 for completely different", func(t *testing.T) {
		s := terlik.LevenshteinSimilarity("abc", "xyz")
		if math.Abs(s) > 0.1 {
			t.Errorf("got %f, want ~0", s)
		}
	})
	t.Run("returns value between 0 and 1", func(t *testing.T) {
		s := terlik.LevenshteinSimilarity("siktir", "siktr")
		if s <= 0.5 || s >= 1.0 {
			t.Errorf("got %f, expected between 0.5 and 1.0", s)
		}
	})
	t.Run("handles empty strings", func(t *testing.T) {
		if s := terlik.LevenshteinSimilarity("", ""); s != 1.0 {
			t.Errorf("got %f, want 1.0", s)
		}
	})
}

func TestDiceSimilarityTests(t *testing.T) {
	t.Run("returns 1 for identical strings", func(t *testing.T) {
		if s := terlik.DiceSimilarity("abc", "abc"); s != 1.0 {
			t.Errorf("got %f, want 1.0", s)
		}
	})
	t.Run("handles single-char strings", func(t *testing.T) {
		if s := terlik.DiceSimilarity("a", "a"); s != 1.0 {
			t.Errorf("a=a: got %f, want 1.0", s)
		}
		if s := terlik.DiceSimilarity("a", "b"); s != 0.0 {
			t.Errorf("a≠b: got %f, want 0.0", s)
		}
	})
	t.Run("returns value between 0 and 1", func(t *testing.T) {
		s := terlik.DiceSimilarity("night", "nacht")
		if s <= 0 || s >= 1.0 {
			t.Errorf("got %f, expected between 0 and 1", s)
		}
	})
	t.Run("returns 0 for no shared bigrams", func(t *testing.T) {
		if s := terlik.DiceSimilarity("ab", "cd"); s != 0.0 {
			t.Errorf("got %f, want 0.0", s)
		}
	})
}

func TestGetFuzzyMatcherTests(t *testing.T) {
	t.Run("returns levenshtein matcher", func(t *testing.T) {
		matcher := terlik.GetFuzzyMatcher(terlik.FuzzyLevenshtein)
		if matcher("abc", "abc") != 1.0 {
			t.Error("expected 1.0 for identical strings")
		}
	})
	t.Run("returns dice matcher", func(t *testing.T) {
		matcher := terlik.GetFuzzyMatcher(terlik.FuzzyDice)
		if matcher("abc", "abc") != 1.0 {
			t.Error("expected 1.0 for identical strings")
		}
	})
}
