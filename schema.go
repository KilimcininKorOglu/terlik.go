package terlik

import (
	"fmt"
	"regexp"
	"slices"
	"strings"
)

var (
	validSeverities = []string{"high", "medium", "low"}
	validCategories = []string{"sexual", "insult", "slur", "general"}
	maxSuffixes     = 150
	suffixPattern   = regexp.MustCompile(`^\p{Ll}{1,10}$`)
)

// rootSet collects lowercased roots for duplicate detection.
func rootSet(entries []DictionaryEntry) map[string]bool {
	set := make(map[string]bool)
	for _, e := range entries {
		set[strings.ToLower(e.Root)] = true
	}
	return set
}

// appendUnique appends values not already tracked in seen, marking them present.
func appendUnique(seen map[string]bool, dst, values []string) []string {
	for _, v := range values {
		if !seen[v] {
			dst = append(dst, v)
			seen[v] = true
		}
	}
	return dst
}

// ValidateDictionary validates raw dictionary data against the expected schema.
func ValidateDictionary(data *DictionaryData) error {
	if data == nil {
		return fmt.Errorf("dictionary data must be a non-null object")
	}
	if data.Version < 1 {
		return fmt.Errorf("dictionary version must be a positive number")
	}
	if err := validateSuffixes(data.Suffixes); err != nil {
		return err
	}
	if err := validateEntries(data.Entries); err != nil {
		return err
	}
	return validateWhitelist(data.Whitelist)
}

// validateSuffixes checks the suffix count limit and the 1-10 lowercase letters format.
func validateSuffixes(suffixes []string) error {
	if len(suffixes) > maxSuffixes {
		return fmt.Errorf("dictionary suffixes exceed maximum of %d", maxSuffixes)
	}
	for _, suffix := range suffixes {
		if !suffixPattern.MatchString(suffix) {
			return fmt.Errorf("invalid suffix %q: must be 1-10 lowercase Unicode letters", suffix)
		}
	}
	return nil
}

// validateEntries checks root presence and uniqueness, severity, and category of each entry.
func validateEntries(entries []DictionaryEntry) error {
	seenRoots := make(map[string]bool)
	for i, entry := range entries {
		label := fmt.Sprintf("entries[%d]", i)

		if len(entry.Root) == 0 {
			return fmt.Errorf("%s: root must be a non-empty string", label)
		}

		rootLower := strings.ToLower(entry.Root)
		if seenRoots[rootLower] {
			return fmt.Errorf("%s: duplicate root %q", label, entry.Root)
		}
		seenRoots[rootLower] = true

		if !slices.Contains(validSeverities, entry.Severity) {
			return fmt.Errorf("%s (root=%q): severity must be one of %s",
				label, entry.Root, strings.Join(validSeverities, ", "))
		}

		if !slices.Contains(validCategories, entry.Category) {
			return fmt.Errorf("%s (root=%q): category must be one of %s",
				label, entry.Root, strings.Join(validCategories, ", "))
		}
	}
	return nil
}

// validateWhitelist rejects empty and duplicate (case-insensitive) whitelist entries.
func validateWhitelist(words []string) error {
	seen := make(map[string]bool)
	for i, w := range words {
		if len(w) == 0 {
			return fmt.Errorf("whitelist[%d]: must not be empty", i)
		}
		wLower := strings.ToLower(w)
		if seen[wLower] {
			return fmt.Errorf("whitelist[%d]: duplicate entry %q", i, w)
		}
		seen[wLower] = true
	}
	return nil
}

// MergeDictionaries merges an extension dictionary into a base dictionary.
// Duplicate roots in the extension are skipped.
func MergeDictionaries(base, ext DictionaryData) DictionaryData {
	existingRoots := rootSet(base.Entries)
	mergedEntries := make([]DictionaryEntry, len(base.Entries))
	copy(mergedEntries, base.Entries)
	for _, entry := range ext.Entries {
		rootLower := strings.ToLower(entry.Root)
		if !existingRoots[rootLower] {
			mergedEntries = append(mergedEntries, entry)
			existingRoots[rootLower] = true
		}
	}

	suffixSet := make(map[string]bool)
	var mergedSuffixes []string
	mergedSuffixes = appendUnique(suffixSet, mergedSuffixes, base.Suffixes)
	mergedSuffixes = appendUnique(suffixSet, mergedSuffixes, ext.Suffixes)

	wlSet := make(map[string]bool)
	var mergedWhitelist []string
	mergedWhitelist = appendUnique(wlSet, mergedWhitelist, base.Whitelist)
	mergedWhitelist = appendUnique(wlSet, mergedWhitelist, ext.Whitelist)

	return DictionaryData{
		Version:   base.Version,
		Suffixes:  mergedSuffixes,
		Entries:   mergedEntries,
		Whitelist: mergedWhitelist,
	}
}
