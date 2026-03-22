package terlik_test

import (
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func TestDetectorPatternMode(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("detects plain profanity", func(t *testing.T) {
		matches := tr.GetMatches("bu adam siktir olsun", nil)
		if len(matches) == 0 {
			t.Fatal("expected matches")
		}
		assertDetectsRoot(t, tr, "bu adam siktir olsun", "sik")
	})
	t.Run("detects leet speak", func(t *testing.T) {
		assertDetects(t, tr, "$1kt1r lan")
	})
	t.Run("detects with separators", func(t *testing.T) {
		assertDetects(t, tr, "s.i.k.t.i.r")
	})
	t.Run("detects with repeated characters", func(t *testing.T) {
		assertDetects(t, tr, "siiiktir")
	})
	t.Run("respects whitelist - sikke should not match", func(t *testing.T) {
		matches := tr.GetMatches("osmanlı sikke koleksiyonu", nil)
		for _, r := range matches {
			if r.Word == "sikke" {
				t.Error("sikke should be whitelisted")
			}
		}
	})
	t.Run("detects orospu", func(t *testing.T) {
		assertDetectsRoot(t, tr, "orospu cocugu", "orospu")
	})
	t.Run("returns empty for clean text", func(t *testing.T) {
		matches := tr.GetMatches("merhaba dunya nasilsin", nil)
		if len(matches) != 0 {
			t.Errorf("expected 0 matches, got %d", len(matches))
		}
	})
}

func TestDetectorStrictMode(t *testing.T) {
	tr := mustNew(t, &terlik.Options{Mode: terlik.ModeStrict})

	t.Run("detects exact matches after normalization", func(t *testing.T) {
		assertDetects(t, tr, "siktir git")
	})
	t.Run("does not detect separated chars", func(t *testing.T) {
		assertClean(t, tr, "s i k t i r")
	})
}

func TestDetectorLooseMode(t *testing.T) {
	t.Run("detects fuzzy matches", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{Mode: terlik.ModeLoose, EnableFuzzy: true, FuzzyThreshold: 0.7})
		assertDetects(t, tr, "siktiir")
	})
}

func TestDetectorGetPatterns(t *testing.T) {
	tr := mustNew(t, nil)
	patterns := tr.GetPatterns()
	if len(patterns) == 0 {
		t.Error("expected non-empty patterns")
	}
	if _, ok := patterns["sik"]; !ok {
		t.Error("expected 'sik' in patterns")
	}
}
