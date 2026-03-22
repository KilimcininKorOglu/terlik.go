package terlik_test

import (
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
	"time"
)

func TestLazyCompilation(t *testing.T) {
	t.Run("constructs quickly without eager compilation", func(t *testing.T) {
		start := time.Now()
		mustNew(t, nil)
		elapsed := time.Since(start)
		if elapsed > 50*time.Millisecond {
			t.Errorf("construction took %v, expected < 50ms", elapsed)
		}
	})

	t.Run("detect compiles lazily and returns correct results", func(t *testing.T) {
		tr := mustNew(t, nil)
		assertDetects(t, tr, "siktir git")
		assertClean(t, tr, "merhaba dunya")
	})

	t.Run("getMatches triggers compilation", func(t *testing.T) {
		tr := mustNew(t, nil)
		matches := tr.GetMatches("siktir git", nil)
		if len(matches) == 0 {
			t.Fatal("expected matches")
		}
		if matches[0].Method != terlik.MethodPattern {
			t.Errorf("expected pattern method, got %q", matches[0].Method)
		}
	})

	t.Run("clean triggers compilation", func(t *testing.T) {
		tr := mustNew(t, nil)
		cleaned := tr.Clean("siktir git", nil)
		if cleaned == "siktir git" {
			t.Error("expected cleaned text")
		}
		assertContains(t, cleaned, "git")
	})

	t.Run("strict mode uses hash lookup not patterns", func(t *testing.T) {
		start := time.Now()
		tr := mustNew(t, &terlik.Options{Mode: terlik.ModeStrict})
		constructTime := time.Since(start)

		start = time.Now()
		tr.ContainsProfanity("siktir", nil)
		detectTime := time.Since(start)

		if constructTime > 50*time.Millisecond {
			t.Errorf("strict construction took %v", constructTime)
		}
		if detectTime > 50*time.Millisecond {
			t.Errorf("strict detect took %v", detectTime)
		}
		assertDetects(t, tr, "siktir")
		assertClean(t, tr, "merhaba")
	})

	t.Run("getPatterns triggers compilation", func(t *testing.T) {
		tr := mustNew(t, nil)
		patterns := tr.GetPatterns()
		if len(patterns) == 0 {
			t.Error("expected patterns")
		}
	})

	t.Run("addWords triggers recompile", func(t *testing.T) {
		tr := mustNew(t, nil)
		assertClean(t, tr, "xyztest123")
		tr.AddWords([]string{"xyztest123"})
		assertDetects(t, tr, "xyztest123")
	})

	t.Run("removeWords triggers recompile", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{CustomList: []string{"xyztest456"}})
		assertDetects(t, tr, "xyztest456")
		tr.RemoveWords([]string{"xyztest456"})
		assertClean(t, tr, "xyztest456")
	})

	t.Run("warmup creates ready instances", func(t *testing.T) {
		cache, err := terlik.Warmup([]string{"tr"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		tr := cache["tr"]
		assertDetects(t, tr, "siktir")
		assertClean(t, tr, "merhaba")
	})
}
