package terlik_test

import (
	"terlik"
	"testing"
)

func TestTerlikContainsProfanity(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("returns true for profane text", func(t *testing.T) {
		assertDetects(t, tr, "siktir git")
	})
	t.Run("returns false for clean text", func(t *testing.T) {
		assertClean(t, tr, "merhaba dunya")
	})
	t.Run("returns false for empty input", func(t *testing.T) {
		if tr.ContainsProfanity("", nil) {
			t.Error("empty input should not be profanity")
		}
	})
}

func TestTerlikGetMatches(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("returns match details", func(t *testing.T) {
		matches := tr.GetMatches("siktir git", nil)
		if len(matches) == 0 {
			t.Fatal("expected at least one match")
		}
		m := matches[0]
		if m.Word == "" {
			t.Error("expected non-empty word")
		}
		if m.Root == "" {
			t.Error("expected non-empty root")
		}
		if m.Severity == "" {
			t.Error("expected non-empty severity")
		}
		if m.Method == "" {
			t.Error("expected non-empty method")
		}
	})
	t.Run("returns empty for clean text", func(t *testing.T) {
		matches := tr.GetMatches("merhaba", nil)
		if len(matches) != 0 {
			t.Errorf("expected 0 matches, got %d", len(matches))
		}
	})
}

func TestTerlikClean(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("masks profanity with stars by default", func(t *testing.T) {
		result := tr.Clean("siktir git", nil)
		assertNotContains(t, result, "siktir")
		assertContains(t, result, "*")
	})
	t.Run("supports partial mask", func(t *testing.T) {
		tp := mustNew(t, &terlik.Options{MaskStyle: terlik.MaskPartial})
		result := tp.Clean("siktir git", nil)
		if result == "siktir git" {
			t.Error("expected text to be cleaned")
		}
	})
	t.Run("supports replace mask", func(t *testing.T) {
		tr2 := mustNew(t, &terlik.Options{MaskStyle: terlik.MaskReplace, ReplaceMask: "[küfür]"})
		result := tr2.Clean("siktir git", nil)
		assertContains(t, result, "[küfür]")
	})
	t.Run("returns clean text unchanged", func(t *testing.T) {
		result := tr.Clean("merhaba dunya", nil)
		if result != "merhaba dunya" {
			t.Errorf("expected unchanged, got %q", result)
		}
	})
}

func TestTerlikAddRemoveWords(t *testing.T) {
	t.Run("adds custom words", func(t *testing.T) {
		tr := mustNew(t, nil)
		assertClean(t, tr, "kodumun")
		tr.AddWords([]string{"kodumun"})
		assertDetects(t, tr, "kodumun")
	})
	t.Run("removes words from dictionary", func(t *testing.T) {
		tr := mustNew(t, nil)
		assertDetects(t, tr, "salak")
		tr.RemoveWords([]string{"salak"})
		assertClean(t, tr, "salak")
	})
}

func TestTerlikModes(t *testing.T) {
	t.Run("strict mode does not catch separated chars", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{Mode: terlik.ModeStrict})
		assertClean(t, tr, "s i k t i r")
	})
	t.Run("balanced mode catches separated chars", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{Mode: terlik.ModeBalanced})
		assertDetects(t, tr, "s.i.k.t.i.r")
	})
	t.Run("loose mode enables fuzzy", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{Mode: terlik.ModeLoose})
		matches := tr.GetMatches("siktiir git", nil)
		if len(matches) == 0 {
			t.Error("expected fuzzy match in loose mode")
		}
	})
}

func TestTerlikCustomOptions(t *testing.T) {
	t.Run("respects custom whitelist", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{Whitelist: []string{"testword"}})
		assertClean(t, tr, "sikke")
	})
	t.Run("respects custom word list", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{CustomList: []string{"hiyar"}})
		assertDetects(t, tr, "bu adam hiyar")
	})
	t.Run("respects maxLength", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{MaxLength: 5})
		assertClean(t, tr, "abcde siktir git")
	})
}

func TestTerlikGetPatterns(t *testing.T) {
	tr := mustNew(t, nil)
	patterns := tr.GetPatterns()
	if len(patterns) == 0 {
		t.Error("expected non-empty patterns map")
	}
}
