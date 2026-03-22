package terlik

import (
	"regexp"
	"sort"
	"strings"
	"sync"
	"time"
	"unicode/utf8"
)

// boolVal safely dereferences a *bool pointer, returning false for nil.
func boolVal(p *bool) bool {
	if p == nil {
		return false
	}
	return *p
}

// detector handles profanity detection with pattern matching, fuzzy matching, and caching.
type detector struct {
	dict                 *dictionary
	patterns             []compiledPattern
	compiled             bool
	cacheKey             string
	normalizedWordSet    map[string]bool
	normalizedWordSlice  []string // sorted slice for deterministic iteration
	normalizedWordToRoot map[string]string
	normalizeFn          func(string) string
	safeNormalizeFn      func(string) string
	locale               string
	charClasses          map[string]string
	mu                   sync.RWMutex
}

// patternCache is a package-level cache shared across detector instances.
var (
	patternCacheMu sync.RWMutex
	patternCache   = make(map[string][]compiledPattern)
)

func newDetector(
	dict *dictionary,
	normalizeFn func(string) string,
	safeNormalizeFn func(string) string,
	locale string,
	charClasses map[string]string,
	cacheKey string,
) *detector {
	d := &detector{
		dict:                dict,
		normalizeFn:         normalizeFn,
		safeNormalizeFn:     safeNormalizeFn,
		locale:              locale,
		charClasses:         charClasses,
		cacheKey:            cacheKey,
		normalizedWordSet:   make(map[string]bool),
		normalizedWordToRoot: make(map[string]string),
	}
	d.buildNormalizedLookup()
	return d
}

func (d *detector) buildNormalizedLookup() {
	d.normalizedWordSet = make(map[string]bool)
	d.normalizedWordToRoot = make(map[string]string)
	d.normalizedWordSlice = nil
	for _, word := range d.dict.getAllWords() {
		n := d.normalizeFn(word)
		if !d.normalizedWordSet[n] {
			d.normalizedWordSlice = append(d.normalizedWordSlice, n)
			d.normalizedWordToRoot[n] = word
		}
		d.normalizedWordSet[n] = true
	}
	// Sort for deterministic fuzzy iteration (Go maps have random iteration order)
	sort.Strings(d.normalizedWordSlice)
}

func (d *detector) ensureCompiled() []compiledPattern {
	d.mu.Lock()
	defer d.mu.Unlock()

	if d.compiled {
		return d.patterns
	}

	if d.cacheKey != "" {
		patternCacheMu.RLock()
		cached, ok := patternCache[d.cacheKey]
		patternCacheMu.RUnlock()
		if ok {
			d.patterns = cached
			d.compiled = true
			return d.patterns
		}
	}

	d.patterns = compilePatterns(
		d.dict.getEntries(),
		d.dict.getSuffixes(),
		d.charClasses,
		d.normalizeFn,
	)
	d.compiled = true

	if d.cacheKey != "" {
		patternCacheMu.Lock()
		patternCache[d.cacheKey] = d.patterns
		patternCacheMu.Unlock()
	}

	return d.patterns
}

func (d *detector) compile() {
	d.ensureCompiled()
}

func (d *detector) recompile() {
	// Compile patterns without lock (expensive, reads only immutable data)
	patterns := compilePatterns(
		d.dict.getEntries(),
		d.dict.getSuffixes(),
		d.charClasses,
		d.normalizeFn,
	)

	// Assign all shared state under a single lock
	d.mu.Lock()
	d.cacheKey = ""
	d.patterns = patterns
	d.compiled = true
	d.buildNormalizedLookup()
	d.mu.Unlock()
}

func (d *detector) getPatterns() map[string]*regexp.Regexp {
	patterns := d.ensureCompiled()
	result := make(map[string]*regexp.Regexp, len(patterns))
	for _, p := range patterns {
		result[p.root] = p.regex
	}
	return result
}

func (d *detector) detect(text string, options *DetectOptions) []MatchResult {
	mode := ModeBalanced
	if options != nil && options.Mode != "" {
		mode = options.Mode
	}

	var results []MatchResult
	whitelist := d.dict.getWhitelist()

	if mode == ModeStrict {
		d.detectStrict(text, whitelist, &results)
	} else {
		d.detectPattern(text, whitelist, &results, options)
	}

	if mode == ModeLoose || (options != nil && boolVal(options.EnableFuzzy)) {
		threshold := 0.8
		algorithm := FuzzyLevenshtein
		if options != nil {
			if options.FuzzyThreshold != 0 {
				threshold = options.FuzzyThreshold
			}
			if options.FuzzyAlgorithm != "" {
				algorithm = options.FuzzyAlgorithm
			}
		}
		d.detectFuzzy(text, whitelist, &results, threshold, algorithm)
	}

	filtered := d.applyStrictnessFilters(results, options)
	return d.deduplicateResults(filtered)
}

func (d *detector) applyStrictnessFilters(results []MatchResult, options *DetectOptions) []MatchResult {
	if options == nil {
		return results
	}

	minSev := options.MinSeverity
	exCats := options.ExcludeCategories
	if minSev == "" && len(exCats) == 0 {
		return results
	}

	var filtered []MatchResult
	for _, r := range results {
		if minSev != "" && SeverityOrder[r.Severity] < SeverityOrder[minSev] {
			continue
		}
		if len(exCats) > 0 && r.Category != "" {
			excluded := false
			for _, c := range exCats {
				if r.Category == c {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}
		filtered = append(filtered, r)
	}
	return filtered
}

func (d *detector) detectStrict(text string, whitelist map[string]bool, results *[]MatchResult) {
	normalized := d.normalizeFn(text)
	words := strings.Fields(normalized)
	origWords, origPositions := fieldPositions(text)

	d.mu.RLock()
	wordSet := d.normalizedWordSet
	wordToRoot := d.normalizedWordToRoot
	d.mu.RUnlock()

	for wi := 0; wi < len(origWords); wi++ {
		origWord := origWords[wi]
		normWord := ""
		if wi < len(words) {
			normWord = words[wi]
		}

		if normWord == "" {
			continue
		}

		if whitelist[normWord] {
			continue
		}

		if wordSet[normWord] {
			dictWord := wordToRoot[normWord]
			entry := d.dict.findRootForWord(dictWord)
			if entry != nil {
				byteIndex := 0
				if wi < len(origPositions) {
					byteIndex = origPositions[wi]
				}
				*results = append(*results, MatchResult{
					Word:     origWord,
					Root:     entry.Root,
					Index:    byteIndex,
					Severity: entry.Severity,
					Category: Category(entry.Category),
					Method:   MethodExact,
				})
			}
		}
	}
}

// fieldPositions splits text on whitespace and returns words with their byte positions.
func fieldPositions(text string) ([]string, []int) {
	var words []string
	var positions []int
	i := 0
	bytes := []byte(text)
	n := len(bytes)

	for i < n {
		// Skip whitespace
		for i < n {
			r, size := utf8.DecodeRune(bytes[i:])
			if !isWhitespaceRune(r) {
				break
			}
			i += size
		}
		if i >= n {
			break
		}
		// Find word end
		start := i
		for i < n {
			r, size := utf8.DecodeRune(bytes[i:])
			if isWhitespaceRune(r) {
				break
			}
			i += size
		}
		words = append(words, text[start:i])
		positions = append(positions, start)
	}
	return words, positions
}

func isWhitespaceRune(r rune) bool {
	return r == ' ' || r == '\t' || r == '\n' || r == '\r' || r == '\f' || r == '\v' ||
		r == 0x00A0 || r == 0x1680 || (r >= 0x2000 && r <= 0x200A) ||
		r == 0x2028 || r == 0x2029 || r == 0x202F || r == 0x205F || r == 0x3000
}

func (d *detector) detectPattern(text string, whitelist map[string]bool, results *[]MatchResult, options *DetectOptions) {
	activeNormFn := d.normalizeFn
	if options != nil && boolVal(options.DisableLeetDecode) {
		activeNormFn = d.safeNormalizeFn
	}

	// Pass 1: locale-lowered text
	lowerText := localeLowerCase(text, d.locale)
	isNorm := lowerText != text && len(lowerText) == len(text)
	d.runPatterns(lowerText, text, whitelist, results, isNorm, options)

	// Pass 2: fully normalized text
	normalizedText := activeNormFn(text)
	if normalizedText != lowerText && len(normalizedText) > 0 {
		d.runPatterns(normalizedText, text, whitelist, results, true, options)
	}

	// Pass 3: CamelCase decompounding
	if options == nil || !boolVal(options.DisableCompound) {
		decompound := camelCaseRe1.ReplaceAllString(text, "${1} ${2}")
		decompound = camelCaseRe2.ReplaceAllString(decompound, "${1} ${2}")
		if decompound != text {
			decompoundNorm := activeNormFn(decompound)
			if decompoundNorm != normalizedText && decompoundNorm != lowerText {
				d.runPatterns(decompoundNorm, text, whitelist, results, true, options)
			}
		}
	}
}

var (
	camelCaseRe1     = regexp.MustCompile(`([a-z])([A-Z])`)
	camelCaseRe2     = regexp.MustCompile(`([A-Z]{2,})([a-z])`)
	endsWithDigitsRe = regexp.MustCompile(`\d+$`)
	nonDigitDigitsRe = regexp.MustCompile(`^[^\d]+\d+$`)
	whitespaceSplitRe = regexp.MustCompile(`\s+`)
)

func (d *detector) runPatterns(
	searchText string,
	originalText string,
	whitelist map[string]bool,
	results *[]MatchResult,
	isNormalized bool,
	options *DetectOptions,
) {
	existingIndices := make(map[int]bool)
	for _, r := range *results {
		existingIndices[r.Index] = true
	}

	patterns := d.ensureCompiled()

	var minSev Severity
	var exCats []Category
	if options != nil {
		minSev = options.MinSeverity
		exCats = options.ExcludeCategories
	}

	callStart := time.Now()

	for _, pattern := range patterns {
		if time.Since(callStart).Milliseconds() > regexTimeoutMs {
			break
		}

		// Skip patterns that will be filtered anyway
		if minSev != "" && SeverityOrder[pattern.severity] < SeverityOrder[minSev] {
			continue
		}
		if len(exCats) > 0 && pattern.category != "" {
			excluded := false
			for _, c := range exCats {
				if pattern.category == c {
					excluded = true
					break
				}
			}
			if excluded {
				continue
			}
		}

		matches := findMatchesWithBoundaries(pattern.regex, searchText)

		for _, m := range matches {
			matchedText := searchText[m[0]:m[1]]
			matchIndex := m[0]

			// Whitelist checks
			if whitelist[matchedText] {
				continue
			}
			normalizedMatch := d.normalizeFn(matchedText)
			if whitelist[normalizedMatch] {
				continue
			}

			surrounding := getSurroundingWord(searchText, matchIndex, len(matchedText))
			if whitelist[surrounding] {
				continue
			}
			normalizedSurrounding := d.normalizeFn(surrounding)
			if whitelist[normalizedSurrounding] {
				continue
			}

			if isNormalized {
				mapped := d.mapNormalizedToOriginal(originalText, matchIndex, matchedText)
				if mapped != nil && whitelist[strings.ToLower(mapped.word)] {
					continue
				}
				// Reject matches where the original word ends with only digits
				if mapped != nil && endsWithDigitsRe.MatchString(mapped.word) && nonDigitDigitsRe.MatchString(mapped.word) {
					continue
				}
				if mapped != nil && !existingIndices[mapped.index] {
					*results = append(*results, MatchResult{
						Word:     mapped.word,
						Root:     pattern.root,
						Index:    mapped.index,
						Severity: pattern.severity,
						Category: pattern.category,
						Method:   MethodPattern,
					})
					existingIndices[mapped.index] = true
				}
			} else {
				if !existingIndices[matchIndex] {
					*results = append(*results, MatchResult{
						Word:     matchedText,
						Root:     pattern.root,
						Index:    matchIndex,
						Severity: pattern.severity,
						Category: pattern.category,
						Method:   MethodPattern,
					})
					existingIndices[matchIndex] = true
				}
			}
		}
	}
}

type mappedWord struct {
	word  string
	index int
}

func (d *detector) mapNormalizedToOriginal(originalText string, normIndex int, _ string) *mappedWord {
	// Split original text preserving whitespace separators
	segments := whitespaceSplitRe.Split(originalText, -1)
	separators := whitespaceSplitRe.FindAllString(originalText, -1)

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
		normEnd := normOffset + len(normWord)

		if normIndex >= normOffset && normIndex < normEnd {
			return &mappedWord{word: segment, index: origOffset}
		}

		normOffset = normEnd
		origOffset += len(segment)

		// Add separator after segment
		if i < len(separators) {
			normOffset++ // normalized whitespace is single space
			origOffset += len(separators[i])
		}
	}

	return nil
}

func (d *detector) detectFuzzy(
	text string,
	whitelist map[string]bool,
	results *[]MatchResult,
	threshold float64,
	algorithm FuzzyAlgorithm,
) {
	normalized := d.normalizeFn(text)
	normWords := strings.Fields(normalized)
	origWords, origPositions := fieldPositions(text)
	matcher := GetFuzzyMatcher(algorithm)

	d.mu.RLock()
	wordSlice := d.normalizedWordSlice
	wordToRoot := d.normalizedWordToRoot
	d.mu.RUnlock()

	existingIndices := make(map[int]bool)
	for _, r := range *results {
		existingIndices[r.Index] = true
	}

	startTime := time.Now()

	for wi := 0; wi < len(origWords); wi++ {
		if time.Since(startTime).Milliseconds() > regexTimeoutMs {
			break
		}

		origWord := origWords[wi]
		word := ""
		if wi < len(normWords) {
			word = normWords[wi]
		}

		if utf8.RuneCountInString(word) < 3 || whitelist[word] {
			continue
		}

		byteIndex := 0
		if wi < len(origPositions) {
			byteIndex = origPositions[wi]
		}

		for _, normDict := range wordSlice {
			if utf8.RuneCountInString(normDict) < 3 {
				continue
			}

			similarity := matcher(word, normDict)
			if similarity >= threshold {
				if !existingIndices[byteIndex] {
					dictWord := wordToRoot[normDict]
					entry := d.dict.findRootForWord(dictWord)
					if entry != nil {
						*results = append(*results, MatchResult{
							Word:     origWord,
							Root:     entry.Root,
							Index:    byteIndex,
							Severity: entry.Severity,
							Category: Category(entry.Category),
							Method:   MethodFuzzy,
						})
						existingIndices[byteIndex] = true
					}
				}
				break
			}
		}
	}
}

// getSurroundingWord expands a match to the full surrounding word.
func getSurroundingWord(text string, index int, length int) string {
	start := index
	end := index + length

	// Walk backward through extended letters
	for start > 0 {
		r, size := utf8.DecodeLastRuneInString(text[:start])
		if !isExtendedLetter(r) {
			break
		}
		start -= size
	}

	// Walk forward through extended letters
	for end < len(text) {
		r, size := utf8.DecodeRuneInString(text[end:])
		if !isExtendedLetter(r) {
			break
		}
		end += size
	}

	return text[start:end]
}

func (d *detector) deduplicateResults(results []MatchResult) []MatchResult {
	seen := make(map[int]MatchResult)
	for _, r := range results {
		existing, ok := seen[r.Index]
		if !ok || utf8.RuneCountInString(r.Word) > utf8.RuneCountInString(existing.Word) {
			seen[r.Index] = r
		}
	}

	var deduped []MatchResult
	for _, r := range seen {
		deduped = append(deduped, r)
	}
	sort.Slice(deduped, func(i, j int) bool {
		return deduped[i].Index < deduped[j].Index
	})
	return deduped
}
