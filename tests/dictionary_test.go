package terlik_test

import (
	"github.com/KilimcininKorOglu/terlik.go"
	"strings"
	"testing"
)

func TestDictionaryJSONSchema(t *testing.T) {
	dict := terlik.TrConfig.Dictionary
	t.Run("validates actual tr.json without errors", func(t *testing.T) {
		if err := terlik.ValidateDictionary(&terlik.TrConfig.Dictionary); err != nil {
			t.Errorf("ValidateDictionary(tr) failed: %v", err)
		}
	})
	t.Run("has valid version", func(t *testing.T) {
		if dict.Version < 1 {
			t.Error("version should be >= 1")
		}
	})
	t.Run("has entries", func(t *testing.T) {
		if len(dict.Entries) == 0 {
			t.Error("expected entries")
		}
	})
	t.Run("has whitelist", func(t *testing.T) {
		if len(dict.Whitelist) == 0 {
			t.Error("expected whitelist")
		}
	})
	t.Run("every entry has non-empty root", func(t *testing.T) {
		assertNoEmptyRoots(t, dict.Entries)
	})
	t.Run("no duplicate roots", func(t *testing.T) {
		assertNoDuplicateRoots(t, dict.Entries)
	})
	t.Run("every entry has valid severity", func(t *testing.T) {
		assertEntryFieldValid(t, dict.Entries, "severity",
			map[string]bool{"high": true, "medium": true, "low": true},
			func(e terlik.DictionaryEntry) string { return e.Severity })
	})
	t.Run("every entry has valid category", func(t *testing.T) {
		assertEntryFieldValid(t, dict.Entries, "category",
			map[string]bool{"sexual": true, "insult": true, "slur": true, "general": true},
			func(e terlik.DictionaryEntry) string { return e.Category })
	})
	t.Run("whitelist contains known safe words", func(t *testing.T) {
		wl := make(map[string]bool)
		for _, w := range dict.Whitelist {
			wl[strings.ToLower(w)] = true
		}
		for _, safe := range []string{"amsterdam", "sikke", "bokser", "malzeme", "memur"} {
			if !wl[safe] {
				t.Errorf("expected %q in whitelist", safe)
			}
		}
	})
}

// assertNoEmptyRoots verifies every dictionary entry carries a root word.
func assertNoEmptyRoots(t *testing.T, entries []terlik.DictionaryEntry) {
	t.Helper()
	for _, e := range entries {
		if e.Root == "" {
			t.Error("found entry with empty root")
		}
	}
}

// assertNoDuplicateRoots verifies entries have unique roots, case-insensitive.
func assertNoDuplicateRoots(t *testing.T, entries []terlik.DictionaryEntry) {
	t.Helper()
	seen := make(map[string]bool)
	for _, e := range entries {
		lower := strings.ToLower(e.Root)
		if seen[lower] {
			t.Errorf("duplicate root: %s", e.Root)
		}
		seen[lower] = true
	}
}

// assertEntryFieldValid verifies the selected entry field holds one of the
// allowed values for every entry.
func assertEntryFieldValid(t *testing.T, entries []terlik.DictionaryEntry, field string, valid map[string]bool, value func(terlik.DictionaryEntry) string) {
	t.Helper()
	for _, e := range entries {
		v := value(e)
		if !valid[v] {
			t.Errorf("invalid %s %q for root %q", field, v, e.Root)
		}
	}
}

func TestValidateDictionaryRejection(t *testing.T) {
	tests := []struct {
		name    string
		dict    *terlik.DictionaryData
		wantSub string
	}{
		{
			name: "rejects nil",
			dict: nil,
		},
		{
			name:    "rejects missing version",
			dict:    &terlik.DictionaryData{Version: 0},
			wantSub: "version",
		},
		{
			name: "rejects duplicate roots",
			dict: &terlik.DictionaryData{
				Version: 1,
				Entries: []terlik.DictionaryEntry{
					{Root: "test", Severity: "high", Category: "general"},
					{Root: "test", Severity: "low", Category: "insult"},
				},
			},
			wantSub: "duplicate",
		},
		{
			name: "rejects invalid severity",
			dict: &terlik.DictionaryData{
				Version: 1,
				Entries: []terlik.DictionaryEntry{
					{Root: "test", Severity: "extreme", Category: "general"},
				},
			},
			wantSub: "severity",
		},
		{
			name: "rejects invalid category",
			dict: &terlik.DictionaryData{
				Version: 1,
				Entries: []terlik.DictionaryEntry{
					{Root: "test", Severity: "high", Category: "unknown"},
				},
			},
			wantSub: "category",
		},
		{
			name: "rejects empty root",
			dict: &terlik.DictionaryData{
				Version: 1,
				Entries: []terlik.DictionaryEntry{
					{Root: "", Severity: "high", Category: "general"},
				},
			},
			wantSub: "root",
		},
		{
			name:    "rejects invalid suffix format",
			dict:    &terlik.DictionaryData{Version: 1, Suffixes: []string{"ABC"}},
			wantSub: "suffix",
		},
		{
			name:    "rejects empty whitelist entry",
			dict:    &terlik.DictionaryData{Version: 1, Whitelist: []string{"valid", ""}},
			wantSub: "empty",
		},
		{
			name:    "rejects duplicate whitelist entry",
			dict:    &terlik.DictionaryData{Version: 1, Whitelist: []string{"word", "word"}},
			wantSub: "duplicate",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := terlik.ValidateDictionary(tt.dict)
			if err == nil {
				t.Fatal("expected error, got nil")
			}
			if tt.wantSub != "" && !strings.Contains(err.Error(), tt.wantSub) {
				t.Errorf("expected error containing %q, got: %v", tt.wantSub, err)
			}
		})
	}
}
