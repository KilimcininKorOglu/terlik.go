package terlik

import (
	"fmt"
	"log"
	"regexp"
	"slices"
	"sort"
	"strings"
	"unicode/utf8"
)

const (
	wordCharRange    = `a-zA-Z0-9\x{00C0}-\x{024F}`
	separator        = `[^` + wordCharRange + `]{0,3}`
	boundaryBefore   = `(?:^|[^` + wordCharRange + `])`
	boundaryAfter    = `(?:[^` + wordCharRange + `]|$)`
	maxPatternLength = 20000
	maxSuffixChain   = 2
	regexTimeoutMs   = 250
)

// escapeRegex escapes regex special characters in a string.
func escapeRegex(s string) string {
	special := `.*+?^${}()|[]\`
	var b strings.Builder
	b.Grow(len(s) * 2)
	for _, r := range s {
		if strings.ContainsRune(special, r) {
			b.WriteRune('\\')
		}
		b.WriteRune(r)
	}
	return b.String()
}

// charToPattern converts a single character to a regex pattern using charClasses.
func charToPattern(ch rune, charClasses map[string]string) string {
	lower := strings.ToLower(string(ch))
	if cls, ok := charClasses[lower]; ok {
		return cls + "+"
	}
	return escapeRegex(string(ch)) + "+"
}

// wordToPattern converts a word to a regex pattern with separators between characters.
func wordToPattern(word string, charClasses map[string]string, normalizeFn func(string) string) string {
	normalized := normalizeFn(word)
	runes := []rune(normalized)
	parts := make([]string, len(runes))
	for i, ch := range runes {
		parts[i] = charToPattern(ch, charClasses)
	}
	return strings.Join(parts, separator)
}

// charToLiteralPattern converts a character to a literal regex pattern (no charClass lookup).
func charToLiteralPattern(ch rune) string {
	return escapeRegex(string(ch)) + "+"
}

// buildSuffixGroup creates a regex group pattern for grammatical suffixes.
func buildSuffixGroup(suffixes []string) string {
	if len(suffixes) == 0 {
		return ""
	}

	suffixPatterns := make([]string, len(suffixes))
	for i, suffix := range suffixes {
		runes := []rune(suffix)
		parts := make([]string, len(runes))
		for j, ch := range runes {
			parts[j] = charToLiteralPattern(ch)
		}
		suffixPatterns[i] = strings.Join(parts, separator)
	}

	// Sort by length descending so longer suffixes match first
	sort.Slice(suffixPatterns, func(i, j int) bool {
		return len(suffixPatterns[i]) > len(suffixPatterns[j])
	})

	return fmt.Sprintf("(?:%s(?:%s))", separator, strings.Join(suffixPatterns, "|"))
}

// sortedEntryRoots returns dictionary keys sorted for deterministic pattern
// order (Go maps have random iteration).
func sortedEntryRoots(entries map[string]WordEntry) []string {
	sortedRoots := make([]string, 0, len(entries))
	for key := range entries {
		sortedRoots = append(sortedRoots, key)
	}
	sort.Strings(sortedRoots)
	return sortedRoots
}

// normalizedForms normalizes root + variants, drops empty and duplicate forms,
// and sorts longest-first so the regex prefers the longest alternative.
func normalizedForms(entry WordEntry, normalizeFn func(string) string) []string {
	allForms := make([]string, 0, 1+len(entry.Variants))
	allForms = append(allForms, entry.Root)
	allForms = append(allForms, entry.Variants...)

	// Normalize, deduplicate, sort by length descending
	seen := make(map[string]bool)
	var sortedForms []string
	for _, w := range allForms {
		n := normalizeFn(w)
		if n == "" || seen[n] {
			continue
		}
		seen[n] = true
		sortedForms = append(sortedForms, n)
	}
	sort.Slice(sortedForms, func(i, j int) bool {
		return utf8.RuneCountInString(sortedForms[i]) > utf8.RuneCountInString(sortedForms[j])
	})
	return sortedForms
}

// joinFormPatterns renders every form as a single regex alternative chain.
func joinFormPatterns(sortedForms []string, charClasses map[string]string, normalizeFn func(string) string) string {
	formPatterns := make([]string, len(sortedForms))
	for i, w := range sortedForms {
		formPatterns[i] = wordToPattern(w, charClasses, normalizeFn)
	}
	return strings.Join(formPatterns, "|")
}

// plainPattern wraps the plain form alternatives in a case-insensitive group.
func plainPattern(sortedForms []string, charClasses map[string]string, normalizeFn func(string) string) string {
	return fmt.Sprintf("(?i)(?:%s)", joinFormPatterns(sortedForms, charClasses, normalizeFn))
}

// entryPattern assembles one entry's inner regex body before boundary
// wrapping: suffixable entries get the suffix chain on every form,
// non-suffixable entries a strict/optional-suffix split by form length,
// and without a suffix group the plain alternatives remain.
func entryPattern(sortedForms []string, suffixGroup string, useSuffix bool, charClasses map[string]string, normalizeFn func(string) string) string {
	if useSuffix {
		// All forms get suffix chain
		return fmt.Sprintf("(?i)(?:%s)%s{0,%d}",
			joinFormPatterns(sortedForms, charClasses, normalizeFn), suffixGroup, maxSuffixChain)
	}
	if suffixGroup != "" {
		return mixedBoundaryPattern(sortedForms, suffixGroup, charClasses, normalizeFn)
	}
	// No suffix group: just the forms
	return plainPattern(sortedForms, charClasses, normalizeFn)
}

// mixedBoundaryPattern handles non-suffixable entries when a suffix group
// exists: short forms (<4 runes) keep a strict boundary, longer forms get an
// optional suffix chain.
func mixedBoundaryPattern(sortedForms []string, suffixGroup string, charClasses map[string]string, normalizeFn func(string) string) string {
	minVariantSuffixLen := 4
	var strictForms, suffixableForms []string
	for _, w := range sortedForms {
		if utf8.RuneCountInString(w) >= minVariantSuffixLen {
			suffixableForms = append(suffixableForms, wordToPattern(w, charClasses, normalizeFn))
		} else {
			strictForms = append(strictForms, wordToPattern(w, charClasses, normalizeFn))
		}
	}
	var parts []string
	if len(suffixableForms) > 0 {
		parts = append(parts, fmt.Sprintf("(?:%s)%s{0,%d}",
			strings.Join(suffixableForms, "|"), suffixGroup, maxSuffixChain))
	}
	if len(strictForms) > 0 {
		parts = append(parts, fmt.Sprintf("(?:%s)", strings.Join(strictForms, "|")))
	}
	return fmt.Sprintf("(?i)(?:%s)", strings.Join(parts, "|"))
}

// compileEntryRegex wraps the inner pattern with boundary assertions via a
// capturing group and compiles it. A suffixed pattern that fails to compile
// falls back to the plain, suffix-free pattern; nil means the entry is
// skipped (both failures are logged).
func compileEntryRegex(root string, pattern string, useSuffix bool, sortedForms []string, charClasses map[string]string, normalizeFn func(string) string) *regexp.Regexp {
	// Wrap with boundary assertions using capturing group.
	// The boundary chars are consumed by the regex but the capturing group
	// extracts only the inner match. This replaces JS lookbehind/lookahead
	// and correctly prevents suffix groups from extending into the next word.
	wrappedPattern := fmt.Sprintf("%s(%s)%s", boundaryBefore, pattern, boundaryAfter)

	re, err := regexp.Compile(wrappedPattern)
	if err == nil {
		return re
	}

	// Fallback: try without suffix
	if !useSuffix {
		log.Printf("[terlik] Pattern for %q failed, skipping: %v", root, err)
		return nil
	}
	fallbackInner := plainPattern(sortedForms, charClasses, normalizeFn)
	fallbackPattern := fmt.Sprintf("%s(%s)%s", boundaryBefore, fallbackInner, boundaryAfter)
	re2, err2 := regexp.Compile(fallbackPattern)
	if err2 != nil {
		log.Printf("[terlik] Pattern for %q failed completely, skipping: %v", root, err2)
		return nil
	}
	log.Printf("[terlik] Pattern for %q failed with suffixes, using fallback: %v", root, err)
	return re2
}

// compilePatterns compiles dictionary entries into regex patterns for detection.
// Patterns are compiled WITHOUT lookbehind/lookahead; word boundary checks are
// performed post-match since Go's RE2 engine doesn't support lookarounds.
func compilePatterns(
	entries map[string]WordEntry,
	suffixes []string,
	charClasses map[string]string,
	normalizeFn func(string) string,
) []compiledPattern {
	var patterns []compiledPattern

	suffixGroup := ""
	if len(suffixes) > 0 {
		suffixGroup = buildSuffixGroup(suffixes)
	}

	for _, key := range sortedEntryRoots(entries) {
		entry := entries[key]
		sortedForms := normalizedForms(entry, normalizeFn)
		if len(sortedForms) == 0 {
			continue
		}

		useSuffix := entry.Suffixable && suffixGroup != ""

		pattern := entryPattern(sortedForms, suffixGroup, useSuffix, charClasses, normalizeFn)

		// Safety guard: if pattern is too long, fallback to non-suffix version
		if len(pattern) > maxPatternLength {
			pattern = plainPattern(sortedForms, charClasses, normalizeFn)
		}

		re := compileEntryRegex(entry.Root, pattern, useSuffix, sortedForms, charClasses, normalizeFn)
		if re == nil {
			continue
		}

		patterns = append(patterns, compiledPattern{
			root:     entry.Root,
			severity: entry.Severity,
			category: Category(entry.Category),
			regex:    re,
			variants: entry.Variants,
		})
	}

	return patterns
}

// findMatchesWithBoundaries uses the compiled regex (which includes boundary
// assertions via capturing groups) to find matches. The regex pattern is:
//
//	(?:^|[^WORD_CHAR])(inner_pattern)(?:[^WORD_CHAR]|$)
//
// Group 1 captures the actual match without boundary chars.
// We iterate manually to handle overlapping matches (where boundary chars
// from one match overlap with the next match's content).
func findMatchesWithBoundaries(re *regexp.Regexp, text string) [][2]int {
	var result [][2]int
	pos := 0

	for pos <= len(text) {
		loc := re.FindStringSubmatchIndex(text[pos:])
		if loc == nil {
			break
		}

		// loc[0:2] = full match, loc[2:4] = group 1 (inner match)
		if len(loc) < 4 || loc[2] < 0 {
			break
		}

		innerStart := pos + loc[2]
		innerEnd := pos + loc[3]

		if innerEnd > innerStart {
			result = append(result, [2]int{innerStart, innerEnd})
		}

		// Advance past the inner match start to find overlapping matches.
		// We advance by at least 1 byte to avoid infinite loops.
		if loc[2] > 0 {
			pos += loc[2]
		} else {
			_, size := utf8.DecodeRuneInString(text[pos:])
			if size == 0 {
				break
			}
			pos += size
		}
	}

	return result
}

// normMapping precomputes the normalized-offset → original-segment table for
// one original text so each match maps back in O(log n) instead of re-splitting
// and re-normalizing the whole text per match.
type normMapping struct {
	starts []int // normalized byte offset where each kept segment begins
	lens   []int // normalized byte length of each kept segment
	words  []string
	origs  []int // original byte offset of each kept segment
}

// lookup returns the original word containing the given normalized byte
// offset. Same semantics as the former linear walk: segments map disjoint
// normalized ranges, invisible-only segments are skipped.
func (nm *normMapping) lookup(normIndex int) (string, int, bool) {
	if nm == nil {
		return "", 0, false
	}
	i, _ := slices.BinarySearch(nm.starts, normIndex+1) // last start <= normIndex
	if i == 0 {
		return "", 0, false
	}
	i--
	if normIndex >= nm.starts[i]+nm.lens[i] {
		return "", 0, false
	}
	return nm.words[i], nm.origs[i], true
}

// buildNormMapping splits and normalizes the original text once, recording
// the normalized range of every segment that normalizes to non-empty text.
func (d *detector) buildNormMapping(originalText string) *normMapping {
	segments := whitespaceSplitRe.Split(originalText, -1)
	separators := whitespaceSplitRe.FindAllString(originalText, -1)

	nm := &normMapping{}
	normOffset := 0
	origOffset := 0
	for i, segment := range segments {
		if segment == "" {
			if i < len(separators) {
				origOffset += len(separators[i])
			}
			continue
		}

		normWord := d.normalizeFn(segment)
		if len(normWord) == 0 {
			origOffset += len(segment)
			if i < len(separators) {
				origOffset += len(separators[i])
			}
			continue
		}

		nm.starts = append(nm.starts, normOffset)
		nm.lens = append(nm.lens, len(normWord))
		nm.words = append(nm.words, segment)
		nm.origs = append(nm.origs, origOffset)
		normOffset += len(normWord)
		origOffset += len(segment)
		if i < len(separators) {
			normOffset++ // normalized whitespace is single space
			origOffset += len(separators[i])
		}
	}
	return nm
}
