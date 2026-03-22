package terlik_test

import (
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func TestMultiLanguageIsolation(t *testing.T) {
	tr := mustNew(t, &terlik.Options{Language: "tr"})
	en := mustNew(t, &terlik.Options{Language: "en"})
	es := mustNew(t, &terlik.Options{Language: "es"})
	de := mustNew(t, &terlik.Options{Language: "de"})

	t.Run("Turkish detects Turkish not others", func(t *testing.T) {
		assertDetects(t, tr, "siktir git")
		assertClean(t, tr, "what the fuck")
		assertClean(t, tr, "mierda")
		assertClean(t, tr, "scheiße")
	})
	t.Run("English detects English not others", func(t *testing.T) {
		assertDetects(t, en, "what the fuck")
		assertClean(t, en, "siktir git")
		assertClean(t, en, "mierda")
		assertClean(t, en, "scheiße")
	})
	t.Run("Spanish detects Spanish not others", func(t *testing.T) {
		assertDetects(t, es, "mierda")
		assertClean(t, es, "siktir git")
		assertClean(t, es, "what the fuck")
		assertClean(t, es, "scheiße")
	})
	t.Run("German detects German not others", func(t *testing.T) {
		assertDetects(t, de, "scheiße")
		assertClean(t, de, "siktir git")
		assertClean(t, de, "what the fuck")
		assertClean(t, de, "mierda")
	})
	t.Run("addWords is instance-scoped", func(t *testing.T) {
		en2 := mustNew(t, &terlik.Options{Language: "en"})
		en2.AddWords([]string{"foobar"})
		assertDetects(t, en2, "foobar")
		assertClean(t, en, "foobar")
		assertClean(t, tr, "foobar")
	})
	t.Run("language property is readable", func(t *testing.T) {
		if tr.Language() != "tr" {
			t.Error("expected tr")
		}
		if en.Language() != "en" {
			t.Error("expected en")
		}
		if es.Language() != "es" {
			t.Error("expected es")
		}
		if de.Language() != "de" {
			t.Error("expected de")
		}
	})
	t.Run("default language is Turkish", func(t *testing.T) {
		def := mustNew(t, nil)
		if def.Language() != "tr" {
			t.Error("expected default tr")
		}
		assertDetects(t, def, "siktir")
	})
}

func TestWarmupTests(t *testing.T) {
	t.Run("creates instances for all specified languages", func(t *testing.T) {
		cache, err := terlik.Warmup([]string{"tr", "en", "es", "de"}, nil)
		if err != nil {
			t.Fatal(err)
		}
		if len(cache) != 4 {
			t.Errorf("expected 4 instances, got %d", len(cache))
		}
	})
	t.Run("each instance works independently", func(t *testing.T) {
		cache, _ := terlik.Warmup([]string{"tr", "en"}, nil)
		assertDetects(t, cache["tr"], "siktir")
		assertDetects(t, cache["en"], "fuck")
		assertClean(t, cache["tr"], "fuck")
		assertClean(t, cache["en"], "siktir")
	})
	t.Run("passes base options to all instances", func(t *testing.T) {
		cache, _ := terlik.Warmup([]string{"tr", "en"}, &terlik.Options{Mode: terlik.ModeStrict})
		assertClean(t, cache["tr"], "s i k t i r")
		assertClean(t, cache["en"], "f u c k")
	})
	t.Run("errors for unsupported language", func(t *testing.T) {
		_, err := terlik.Warmup([]string{"tr", "xx"}, nil)
		if err == nil {
			t.Error("expected error for unsupported language")
		}
	})
}
