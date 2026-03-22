package terlik_test

import (
	"strings"
	"terlik"
	"testing"
)

func TestDictionaryJSONSchema(t *testing.T) {
	t.Run("validates actual tr.json without errors", func(t *testing.T) {
		if err := terlik.ValidateDictionary(&terlik.TrConfig.Dictionary); err != nil {
			t.Errorf("ValidateDictionary(tr) failed: %v", err)
		}
	})
	t.Run("has valid version", func(t *testing.T) {
		if terlik.TrConfig.Dictionary.Version < 1 {
			t.Error("version should be >= 1")
		}
	})
	t.Run("has entries", func(t *testing.T) {
		if len(terlik.TrConfig.Dictionary.Entries) == 0 {
			t.Error("expected entries")
		}
	})
	t.Run("has whitelist", func(t *testing.T) {
		if len(terlik.TrConfig.Dictionary.Whitelist) == 0 {
			t.Error("expected whitelist")
		}
	})
	t.Run("every entry has non-empty root", func(t *testing.T) {
		for _, e := range terlik.TrConfig.Dictionary.Entries {
			if e.Root == "" {
				t.Error("found entry with empty root")
			}
		}
	})
	t.Run("no duplicate roots", func(t *testing.T) {
		seen := make(map[string]bool)
		for _, e := range terlik.TrConfig.Dictionary.Entries {
			lower := strings.ToLower(e.Root)
			if seen[lower] {
				t.Errorf("duplicate root: %s", e.Root)
			}
			seen[lower] = true
		}
	})
	t.Run("every entry has valid severity", func(t *testing.T) {
		valid := map[string]bool{"high": true, "medium": true, "low": true}
		for _, e := range terlik.TrConfig.Dictionary.Entries {
			if !valid[e.Severity] {
				t.Errorf("invalid severity %q for root %q", e.Severity, e.Root)
			}
		}
	})
	t.Run("every entry has valid category", func(t *testing.T) {
		valid := map[string]bool{"sexual": true, "insult": true, "slur": true, "general": true}
		for _, e := range terlik.TrConfig.Dictionary.Entries {
			if !valid[e.Category] {
				t.Errorf("invalid category %q for root %q", e.Category, e.Root)
			}
		}
	})
	t.Run("whitelist contains known safe words", func(t *testing.T) {
		wl := make(map[string]bool)
		for _, w := range terlik.TrConfig.Dictionary.Whitelist {
			wl[strings.ToLower(w)] = true
		}
		for _, safe := range []string{"amsterdam", "sikke", "bokser", "malzeme", "memur"} {
			if !wl[safe] {
				t.Errorf("expected %q in whitelist", safe)
			}
		}
	})
}

func TestValidateDictionaryRejection(t *testing.T) {
	t.Run("rejects nil", func(t *testing.T) {
		if err := terlik.ValidateDictionary(nil); err == nil {
			t.Error("expected error for nil")
		}
	})
	t.Run("rejects missing version", func(t *testing.T) {
		err := terlik.ValidateDictionary(&terlik.DictionaryData{Version: 0})
		if err == nil || !strings.Contains(err.Error(), "version") {
			t.Errorf("expected version error, got: %v", err)
		}
	})
	t.Run("rejects duplicate roots", func(t *testing.T) {
		err := terlik.ValidateDictionary(&terlik.DictionaryData{
			Version: 1,
			Entries: []terlik.DictionaryEntry{
				{Root: "test", Severity: "high", Category: "general"},
				{Root: "test", Severity: "low", Category: "insult"},
			},
		})
		if err == nil || !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("expected duplicate error, got: %v", err)
		}
	})
	t.Run("rejects invalid severity", func(t *testing.T) {
		err := terlik.ValidateDictionary(&terlik.DictionaryData{
			Version: 1,
			Entries: []terlik.DictionaryEntry{
				{Root: "test", Severity: "extreme", Category: "general"},
			},
		})
		if err == nil || !strings.Contains(err.Error(), "severity") {
			t.Errorf("expected severity error, got: %v", err)
		}
	})
	t.Run("rejects invalid category", func(t *testing.T) {
		err := terlik.ValidateDictionary(&terlik.DictionaryData{
			Version: 1,
			Entries: []terlik.DictionaryEntry{
				{Root: "test", Severity: "high", Category: "unknown"},
			},
		})
		if err == nil || !strings.Contains(err.Error(), "category") {
			t.Errorf("expected category error, got: %v", err)
		}
	})
	t.Run("rejects empty root", func(t *testing.T) {
		err := terlik.ValidateDictionary(&terlik.DictionaryData{
			Version: 1,
			Entries: []terlik.DictionaryEntry{
				{Root: "", Severity: "high", Category: "general"},
			},
		})
		if err == nil || !strings.Contains(err.Error(), "root") {
			t.Errorf("expected root error, got: %v", err)
		}
	})
	t.Run("rejects invalid suffix format", func(t *testing.T) {
		err := terlik.ValidateDictionary(&terlik.DictionaryData{
			Version:  1,
			Suffixes: []string{"ABC"},
		})
		if err == nil || !strings.Contains(strings.ToLower(err.Error()), "suffix") {
			t.Errorf("expected suffix error, got: %v", err)
		}
	})
	t.Run("rejects empty whitelist entry", func(t *testing.T) {
		err := terlik.ValidateDictionary(&terlik.DictionaryData{
			Version:   1,
			Whitelist: []string{"valid", ""},
		})
		if err == nil || !strings.Contains(err.Error(), "empty") {
			t.Errorf("expected empty error, got: %v", err)
		}
	})
	t.Run("rejects duplicate whitelist entry", func(t *testing.T) {
		err := terlik.ValidateDictionary(&terlik.DictionaryData{
			Version:   1,
			Whitelist: []string{"word", "word"},
		})
		if err == nil || !strings.Contains(err.Error(), "duplicate") {
			t.Errorf("expected duplicate error, got: %v", err)
		}
	})
}
