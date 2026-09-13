package terlik_test

import (
	"github.com/KilimcininKorOglu/terlik.go"
	"strings"
	"testing"
)

func TestMergeDictionariesTests(t *testing.T) {
	base := terlik.DictionaryData{
		Version:   1,
		Suffixes:  []string{"ler", "lar"},
		Entries:   []terlik.DictionaryEntry{{Root: "kötü", Severity: "low", Category: "insult", Suffixable: true}},
		Whitelist: []string{"safeword"},
	}
	ext := terlik.DictionaryData{
		Version:   1,
		Suffixes:  []string{"ler", "ci"},
		Entries:   []terlik.DictionaryEntry{{Root: "badword", Variants: []string{"b4dword"}, Severity: "high", Category: "general"}},
		Whitelist: []string{"safeword", "anothersafe"},
	}

	t.Run("merges entries from extension", func(t *testing.T) {
		merged := terlik.MergeDictionaries(base, ext)
		assertHasRoot(t, merged.Entries, "kötü")
		assertHasRoot(t, merged.Entries, "badword")
	})
	t.Run("skips duplicate roots case-insensitive", func(t *testing.T) {
		extDup := terlik.DictionaryData{
			Version: 1, Entries: []terlik.DictionaryEntry{{Root: "Kötü", Severity: "high", Category: "insult"}},
		}
		merged := terlik.MergeDictionaries(base, extDup)
		count := 0
		for _, e := range merged.Entries {
			if strings.EqualFold(e.Root, "kötü") {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected 1 kötü entry, got %d", count)
		}
	})
	t.Run("unions suffixes deduplicated", func(t *testing.T) {
		merged := terlik.MergeDictionaries(base, ext)
		count := 0
		for _, s := range merged.Suffixes {
			if s == "ler" {
				count++
			}
		}
		if count != 1 {
			t.Errorf("expected 1 'ler' suffix, got %d", count)
		}
	})
	t.Run("preserves base version", func(t *testing.T) {
		merged := terlik.MergeDictionaries(base, ext)
		if merged.Version != base.Version {
			t.Errorf("expected version %d, got %d", base.Version, merged.Version)
		}
	})
}

// assertHasRoot verifies the merged entries contain the given root word.
func assertHasRoot(t *testing.T, entries []terlik.DictionaryEntry, root string) {
	t.Helper()
	for _, e := range entries {
		if e.Root == root {
			return
		}
	}
	t.Errorf("expected root %q in merged entries", root)
}

func TestExtendDictionaryOption(t *testing.T) {
	t.Run("detects words from extended dictionary", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{
			ExtendDictionary: &terlik.DictionaryData{
				Version: 1, Entries: []terlik.DictionaryEntry{
					{Root: "customcurse", Severity: "high", Category: "general"},
				},
			},
		})
		assertDetects(t, tr, "customcurse")
	})
	t.Run("still detects built-in words", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{
			ExtendDictionary: &terlik.DictionaryData{
				Version: 1, Entries: []terlik.DictionaryEntry{
					{Root: "customcurse", Severity: "high", Category: "general"},
				},
			},
		})
		assertDetects(t, tr, "siktir")
	})
	t.Run("works with customList simultaneously", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{
			CustomList: []string{"extraword"},
			ExtendDictionary: &terlik.DictionaryData{
				Version: 1, Entries: []terlik.DictionaryEntry{
					{Root: "extcurse", Severity: "high", Category: "general"},
				},
			},
		})
		assertDetects(t, tr, "extraword")
		assertDetects(t, tr, "extcurse")
		assertDetects(t, tr, "siktir")
	})
	t.Run("rejects invalid extendDictionary", func(t *testing.T) {
		_, err := terlik.New(&terlik.Options{
			ExtendDictionary: &terlik.DictionaryData{Version: -1},
		})
		if err == nil || !strings.Contains(err.Error(), "version") {
			t.Errorf("expected version error, got: %v", err)
		}
	})
	t.Run("disables pattern cache", func(t *testing.T) {
		a := mustNew(t, nil)
		a.ContainsProfanity("warmup", nil)
		b := mustNew(t, &terlik.Options{
			ExtendDictionary: &terlik.DictionaryData{
				Version: 1, Entries: []terlik.DictionaryEntry{
					{Root: "xyznotreal", Severity: "high", Category: "general"},
				},
			},
		})
		assertDetects(t, b, "xyznotreal")
		assertClean(t, a, "xyznotreal")
	})
	t.Run("extended suffixes work for suffixable entries", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{
			ExtendDictionary: &terlik.DictionaryData{
				Version:  1,
				Suffixes: []string{"ler", "lar"},
				Entries:  []terlik.DictionaryEntry{{Root: "extword", Severity: "high", Category: "general", Suffixable: true}},
			},
		})
		assertDetects(t, tr, "extword")
		assertDetects(t, tr, "extwordler")
	})
}
