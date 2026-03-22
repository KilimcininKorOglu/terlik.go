package terlik_test

import (
	"strings"
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func TestLanguageRegistry(t *testing.T) {
	t.Run("returns config for all supported languages", func(t *testing.T) {
		for _, lang := range terlik.GetSupportedLanguages() {
			config, err := terlik.GetLanguageConfig(lang)
			if err != nil {
				t.Errorf("GetLanguageConfig(%q) error: %v", lang, err)
				continue
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
		langs := terlik.GetSupportedLanguages()
		expected := []string{"tr", "en", "es", "de"}
		for _, e := range expected {
			found := false
			for _, l := range langs {
				if l == e {
					found = true
					break
				}
			}
			if !found {
				t.Errorf("expected %q in supported languages", e)
			}
		}
		if len(langs) != 4 {
			t.Errorf("expected 4 languages, got %d", len(langs))
		}
	})

	t.Run("each config has valid charClasses with a-z keys", func(t *testing.T) {
		for _, lang := range terlik.GetSupportedLanguages() {
			config, _ := terlik.GetLanguageConfig(lang)
			for _, key := range []string{"a", "s", "t"} {
				if _, ok := config.CharClasses[key]; !ok {
					t.Errorf("%s: missing charClass key %q", lang, key)
				}
			}
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
