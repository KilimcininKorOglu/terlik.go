package terlik

import (
	"fmt"
	"regexp"
	"strings"
)

var (
	validSeverities = []string{"high", "medium", "low"}
	validCategories = []string{"sexual", "insult", "slur", "general"}
	maxSuffixes     = 150
	suffixPattern   = regexp.MustCompile(`^\p{Ll}{1,10}$`)
)

func containsStr(slice []string, s string) bool {
	for _, v := range slice {
		if v == s {
			return true
		}
	}
	return false
}

// ValidateDictionary validates raw dictionary data against the expected schema.
func ValidateDictionary(data *DictionaryData) error {
	if data == nil {
		return fmt.Errorf("dictionary data must be a non-null object")
	}

	if data.Version < 1 {
		return fmt.Errorf("dictionary version must be a positive number")
	}

	if len(data.Suffixes) > maxSuffixes {
		return fmt.Errorf("dictionary suffixes exceed maximum of %d", maxSuffixes)
	}

	for _, suffix := range data.Suffixes {
		if !suffixPattern.MatchString(suffix) {
			return fmt.Errorf("invalid suffix %q: must be 1-10 lowercase Unicode letters", suffix)
		}
	}

	seenRoots := make(map[string]bool)
	for i, entry := range data.Entries {
		label := fmt.Sprintf("entries[%d]", i)

		if len(entry.Root) == 0 {
			return fmt.Errorf("%s: root must be a non-empty string", label)
		}

		rootLower := strings.ToLower(entry.Root)
		if seenRoots[rootLower] {
			return fmt.Errorf("%s: duplicate root %q", label, entry.Root)
		}
		seenRoots[rootLower] = true

		if !containsStr(validSeverities, entry.Severity) {
			return fmt.Errorf("%s (root=%q): severity must be one of %s",
				label, entry.Root, strings.Join(validSeverities, ", "))
		}

		if !containsStr(validCategories, entry.Category) {
			return fmt.Errorf("%s (root=%q): category must be one of %s",
				label, entry.Root, strings.Join(validCategories, ", "))
		}
	}

	seenWhitelist := make(map[string]bool)
	for i, w := range data.Whitelist {
		if len(w) == 0 {
			return fmt.Errorf("whitelist[%d]: must not be empty", i)
		}
		wLower := strings.ToLower(w)
		if seenWhitelist[wLower] {
			return fmt.Errorf("whitelist[%d]: duplicate entry %q", i, w)
		}
		seenWhitelist[wLower] = true
	}

	return nil
}

// MergeDictionaries merges an extension dictionary into a base dictionary.
// Duplicate roots in the extension are skipped.
func MergeDictionaries(base, ext DictionaryData) DictionaryData {
	existingRoots := make(map[string]bool)
	for _, e := range base.Entries {
		existingRoots[strings.ToLower(e.Root)] = true
	}

	mergedEntries := make([]DictionaryEntry, len(base.Entries))
	copy(mergedEntries, base.Entries)

	for _, entry := range ext.Entries {
		if !existingRoots[strings.ToLower(entry.Root)] {
			mergedEntries = append(mergedEntries, entry)
			existingRoots[strings.ToLower(entry.Root)] = true
		}
	}

	suffixSet := make(map[string]bool)
	var mergedSuffixes []string
	for _, s := range base.Suffixes {
		if !suffixSet[s] {
			mergedSuffixes = append(mergedSuffixes, s)
			suffixSet[s] = true
		}
	}
	for _, s := range ext.Suffixes {
		if !suffixSet[s] {
			mergedSuffixes = append(mergedSuffixes, s)
			suffixSet[s] = true
		}
	}

	wlSet := make(map[string]bool)
	var mergedWhitelist []string
	for _, w := range base.Whitelist {
		if !wlSet[w] {
			mergedWhitelist = append(mergedWhitelist, w)
			wlSet[w] = true
		}
	}
	for _, w := range ext.Whitelist {
		if !wlSet[w] {
			mergedWhitelist = append(mergedWhitelist, w)
			wlSet[w] = true
		}
	}

	return DictionaryData{
		Version:   base.Version,
		Suffixes:  mergedSuffixes,
		Entries:   mergedEntries,
		Whitelist: mergedWhitelist,
	}
}
