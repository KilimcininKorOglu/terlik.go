package terlik

import "strings"

// dictionary manages the profanity word list, whitelist, and suffixes.
type dictionary struct {
	entries   map[string]WordEntry
	whitelist map[string]bool
	allWords  []string
	suffixes  []string
}

// newDictionary creates a new dictionary from validated dictionary data.
func newDictionary(data DictionaryData, customWords []string, customWhitelist []string) *dictionary {
	d := &dictionary{
		entries:   make(map[string]WordEntry),
		whitelist: make(map[string]bool),
	}

	for _, w := range data.Whitelist {
		d.whitelist[strings.ToLower(w)] = true
	}

	d.suffixes = data.Suffixes

	for _, cw := range customWhitelist {
		d.whitelist[strings.ToLower(cw)] = true
	}

	for _, entry := range data.Entries {
		d.addEntry(WordEntry{
			Root:       entry.Root,
			Variants:   entry.Variants,
			Severity:   Severity(entry.Severity),
			Category:   entry.Category,
			Suffixable: entry.Suffixable,
		})
	}

	for _, word := range customWords {
		lower := strings.ToLower(strings.TrimSpace(word))
		if lower == "" {
			continue
		}
		d.addEntry(WordEntry{
			Root:     lower,
			Severity: SeverityMedium,
		})
	}

	return d
}

func (d *dictionary) addEntry(entry WordEntry) {
	normalizedRoot := strings.ToLower(entry.Root)
	d.entries[normalizedRoot] = entry
	d.allWords = append(d.allWords, normalizedRoot)
	for _, v := range entry.Variants {
		d.allWords = append(d.allWords, strings.ToLower(v))
	}
}

func (d *dictionary) getEntries() map[string]WordEntry {
	return d.entries
}

func (d *dictionary) getAllWords() []string {
	return d.allWords
}

func (d *dictionary) getWhitelist() map[string]bool {
	return d.whitelist
}

func (d *dictionary) getSuffixes() []string {
	return d.suffixes
}

func (d *dictionary) addWords(words []string) {
	for _, word := range words {
		lower := strings.ToLower(strings.TrimSpace(word))
		if lower == "" {
			continue
		}
		if _, exists := d.entries[lower]; !exists {
			d.addEntry(WordEntry{
				Root:     lower,
				Severity: SeverityMedium,
			})
		}
	}
}

func (d *dictionary) removeWords(words []string) {
	for _, word := range words {
		key := strings.ToLower(word)
		entry, exists := d.entries[key]
		if !exists {
			continue
		}
		delete(d.entries, key)

		// Build set of words to remove
		removeSet := make(map[string]bool)
		removeSet[key] = true
		for _, v := range entry.Variants {
			removeSet[strings.ToLower(v)] = true
		}

		// Filter allWords (fresh slice to allow GC of removed strings)
		filtered := make([]string, 0, len(d.allWords))
		for _, w := range d.allWords {
			if !removeSet[w] {
				filtered = append(filtered, w)
			}
		}
		d.allWords = filtered
	}
}

func (d *dictionary) findRootForWord(word string) *WordEntry {
	lower := strings.ToLower(word)
	if entry, ok := d.entries[lower]; ok {
		return &entry
	}
	for _, entry := range d.entries {
		for _, v := range entry.Variants {
			if strings.ToLower(v) == lower {
				e := entry
				return &e
			}
		}
	}
	return nil
}
