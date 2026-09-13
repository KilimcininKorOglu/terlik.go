package terlik_test

import (
	"slices"
	"strings"
	"testing"

	terlik "github.com/KilimcininKorOglu/terlik.go"
)

func TestLanguageRegistry(t *testing.T) {
	t.Run("returns config for all supported languages", func(t *testing.T) {
		for _, lang := range terlik.GetSupportedLanguages() {
			assertLanguageConfig(t, lang)
		}
	})

	t.Run("throws for unsupported language", func(t *testing.T) {
		_, err := terlik.GetLanguageConfig("xx")
		if err == nil {
			t.Error("expected error for unsupported language")
		}
	})

	t.Run("error message lists available languages", func(t *testing.T) {
		_, err := terlik.GetLanguageConfig("xx")
		if err == nil || !strings.Contains(err.Error(), "tr") {
			t.Error("expected error to mention available languages")
		}
	})

	t.Run("getSupportedLanguages returns all 4", func(t *testing.T) {
		assertSupportedLanguages(t)
	})

	t.Run("each config has valid charClasses with a-z keys", func(t *testing.T) {
		for _, lang := range terlik.GetSupportedLanguages() {
			assertCharClassKeys(t, lang, "a", "s", "t")
		}
	})

	t.Run("Turkish config has numberExpansions", func(t *testing.T) {
		config, _ := terlik.GetLanguageConfig("tr")
		if len(config.NumberExpansions) == 0 {
			t.Error("expected Turkish numberExpansions")
		}
	})

	t.Run("English config has no numberExpansions", func(t *testing.T) {
		config, _ := terlik.GetLanguageConfig("en")
		if len(config.NumberExpansions) != 0 {
			t.Error("expected no English numberExpansions")
		}
	})
}

// assertLanguageConfig verifies the registry returns a fully populated config
// for the given language.
func assertLanguageConfig(t *testing.T, lang string) {
	t.Helper()
	config, err := terlik.GetLanguageConfig(lang)
	if err != nil {
		t.Errorf("GetLanguageConfig(%q) error: %v", lang, err)
		return
	}
	if config.Locale != lang {
		t.Errorf("expected locale %q, got %q", lang, config.Locale)
	}
	if config.CharMap == nil {
		t.Errorf("%s: charMap is nil", lang)
	}
	if config.LeetMap == nil {
		t.Errorf("%s: leetMap is nil", lang)
	}
	if config.CharClasses == nil {
		t.Errorf("%s: charClasses is nil", lang)
	}
	if len(config.Dictionary.Entries) == 0 {
		t.Errorf("%s: no dictionary entries", lang)
	}
	if config.Dictionary.Version < 1 {
		t.Errorf("%s: dictionary version < 1", lang)
	}
}

// assertSupportedLanguages verifies the registry exposes exactly tr, en, es, de.
func assertSupportedLanguages(t *testing.T) {
	t.Helper()
	langs := terlik.GetSupportedLanguages()
	expected := []string{"tr", "en", "es", "de"}
	for _, e := range expected {
		if !slices.Contains(langs, e) {
			t.Errorf("expected %q in supported languages", e)
		}
	}
	if len(langs) != 4 {
		t.Errorf("expected 4 languages, got %d", len(langs))
	}
}

// assertCharClassKeys verifies the language config defines the given charClass keys.
func assertCharClassKeys(t *testing.T, lang string, keys ...string) {
	t.Helper()
	config, _ := terlik.GetLanguageConfig(lang)
	for _, key := range keys {
		if _, ok := config.CharClasses[key]; !ok {
			t.Errorf("%s: missing charClass key %q", lang, key)
		}
	}
}
