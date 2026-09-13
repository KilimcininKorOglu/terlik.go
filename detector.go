package terlik

import (
	"regexp"
	"slices"
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
		dict:                 dict,
		normalizeFn:          normalizeFn,
		safeNormalizeFn:      safeNormalizeFn,
		locale:               locale,
		charClasses:          charClasses,
		cacheKey:             cacheKey,
		normalizedWordSet:    make(map[string]bool),
		normalizedWordToRoot: make(map[string]string),
	}
	d.buildNormalizedLookup()
	return d
}

// buildNormalizedLookupData builds normalized lookup structures without modifying
// the detector's state. Returns the three structures to be assigned under lock.
func (d *detector) buildNormalizedLookupData() (map[string]bool, []string, map[string]string) {
	wordSet := make(map[string]bool)
	wordToRoot := make(map[string]string)
	var wordSlice []string
	for _, word := range d.dict.getAllWords() {
		n := d.normalizeFn(word)
		if !wordSet[n] {
			wordSlice = append(wordSlice, n)
			wordToRoot[n] = word
		}
		wordSet[n] = true
	}
	sort.Strings(wordSlice)
	return wordSet, wordSlice, wordToRoot
}

// buildNormalizedLookup rebuilds and assigns normalized lookup structures.
// Must be called under d.mu.Lock() or during single-threaded initialization.
func (d *detector) buildNormalizedLookup() {
	d.normalizedWordSet, d.normalizedWordSlice, d.normalizedWordToRoot = d.buildNormalizedLookupData()
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
	// Build everything outside lock (expensive, reads only immutable data)
	patterns := compilePatterns(
		d.dict.getEntries(),
		d.dict.getSuffixes(),
		d.charClasses,
		d.normalizeFn,
	)
	wordSet, wordSlice, wordToRoot := d.buildNormalizedLookupData()

	// Swap all shared state under a short lock
	d.mu.Lock()
	d.cacheKey = ""
	d.patterns = patterns
	d.compiled = true
	d.normalizedWordSet = wordSet
	d.normalizedWordSlice = wordSlice
	d.normalizedWordToRoot = wordToRoot
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

// passesStrictness reports whether a match survives the caller's severity
// floor and category exclusions.
func passesStrictness(r MatchResult, minSev Severity, exCats []Category) bool {
	if minSev != "" && SeverityOrder[r.Severity] < SeverityOrder[minSev] {
		return false
	}
	if len(exCats) > 0 && r.Category != "" && slices.Contains(exCats, r.Category) {
		return false
	}
	return true
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
		if passesStrictness(r, minSev, exCats) {
			filtered = append(filtered, r)
		}
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

	for wi := range origWords {
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

// whitespaceRunes lists every rune treated as a word separator, sorted
// ascending so membership is a deterministic binary search.
var whitespaceRunes = []rune{
	'\t', '\n', '\v', '\f', '\r', ' ',
	0x00A0, 0x1680,
	0x2000, 0x2001, 0x2002, 0x2003, 0x2004, 0x2005, 0x2006, 0x2007, 0x2008, 0x2009, 0x200A,
	0x2028, 0x2029, 0x202F, 0x205F, 0x3000,
}

func isWhitespaceRune(r rune) bool {
	_, found := slices.BinarySearch(whitespaceRunes, r)
	return found
}

type normMapper struct {
	det  *detector
	text string
	once bool
	nm   *normMapping
}

func (m *normMapper) get() *normMapping {
	if !m.once {
		m.nm = m.det.buildNormMapping(m.text)
		m.once = true
	}
	return m.nm
}

func (d *detector) detectPattern(text string, whitelist map[string]bool, results *[]MatchResult, options *DetectOptions) {
	activeNormFn := d.normalizeFn
	if options != nil && boolVal(options.DisableLeetDecode) {
		activeNormFn = d.safeNormalizeFn
	}

	mapper := &normMapper{det: d, text: text}

	lowerText, passOneRan := d.runLocaleLoweredPass(text, whitelist, results, options, mapper)
	normalizedText := d.runNormalizedPass(text, lowerText, passOneRan, activeNormFn, whitelist, results, options, mapper)
	d.runCamelCasePass(text, lowerText, normalizedText, activeNormFn, whitelist, results, options, mapper)
}

// runLocaleLoweredPass runs Pass 1 over the locale-lowered text and reports
// whether it ran. It is skipped when lowering changes the byte length
// (e.g. Turkish I→ı): match positions would no longer map onto the original.
func (d *detector) runLocaleLoweredPass(text string, whitelist map[string]bool, results *[]MatchResult, options *DetectOptions, mapper *normMapper) (string, bool) {
	lowerText := localeLowerCase(text, d.locale)
	if lowerText == text {
		// No case change — positions are identical, no mapping needed
		d.runPatterns(lowerText, text, whitelist, results, false, options, mapper)
		return lowerText, true
	}
	if len(lowerText) == len(text) {
		// Case changed but byte lengths match — safe for normalized mapping
		d.runPatterns(lowerText, text, whitelist, results, true, options, mapper)
		return lowerText, true
	}
	return lowerText, false
}

// runNormalizedPass runs Pass 2 over the fully normalized text — always when
// Pass 1 was skipped, otherwise only when normalization produced new text.
func (d *detector) runNormalizedPass(text string, lowerText string, passOneRan bool, activeNormFn func(string) string, whitelist map[string]bool, results *[]MatchResult, options *DetectOptions, mapper *normMapper) string {
	normalizedText := activeNormFn(text)
	if len(normalizedText) > 0 && (!passOneRan || normalizedText != lowerText) {
		d.runPatterns(normalizedText, text, whitelist, results, true, options, mapper)
	}
	return normalizedText
}

// runCamelCasePass runs Pass 3 over camelCase-decompounded text, provided
// compounding is enabled and decompounding yields text no earlier pass covered.
func (d *detector) runCamelCasePass(text string, lowerText string, normalizedText string, activeNormFn func(string) string, whitelist map[string]bool, results *[]MatchResult, options *DetectOptions, mapper *normMapper) {
	if options != nil && boolVal(options.DisableCompound) {
		return
	}
	decompound := camelCaseRe1.ReplaceAllString(text, "${1} ${2}")
	decompound = camelCaseRe2.ReplaceAllString(decompound, "${1} ${2}")
	if decompound == text {
		return
	}
	decompoundNorm := activeNormFn(decompound)
	if decompoundNorm == normalizedText || decompoundNorm == lowerText {
		return
	}
	d.runPatterns(decompoundNorm, text, whitelist, results, true, options, mapper)
}

var (
	camelCaseRe1      = regexp.MustCompile(`([a-z])([A-Z])`)
	camelCaseRe2      = regexp.MustCompile(`([A-Z]{2,})([a-z])`)
	whitespaceSplitRe = regexp.MustCompile(`\s+`)
)

// buildExistingIndices indexes already-collected results so duplicate byte
// positions are never reported twice.
func buildExistingIndices(results []MatchResult) map[int]bool {
	existingIndices := make(map[int]bool)
	for _, r := range results {
		existingIndices[r.Index] = true
	}
	return existingIndices
}

func (d *detector) runPatterns(
	searchText string,
	originalText string,
	whitelist map[string]bool,
	results *[]MatchResult,
	isNormalized bool,
	options *DetectOptions,
	mapper *normMapper,
) {
	existingIndices := buildExistingIndices(*results)

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
		if !patternPassesFilters(pattern, minSev, exCats) {
			continue
		}

		for _, m := range findMatchesWithBoundaries(pattern.regex, searchText) {
			d.collectPatternMatch(pattern, searchText, originalText, m, whitelist, isNormalized, existingIndices, results, mapper)
		}
	}
}

// patternPassesFilters reports whether a compiled pattern survives the
// caller's severity floor and category exclusions.
func patternPassesFilters(pattern compiledPattern, minSev Severity, exCats []Category) bool {
	if minSev != "" && SeverityOrder[pattern.severity] < SeverityOrder[minSev] {
		return false
	}
	if len(exCats) > 0 && pattern.category != "" && slices.Contains(exCats, pattern.category) {
		return false
	}
	return true
}

// collectPatternMatch runs the whitelist checks for one regex match and
// records it — mapped back onto the original text when the search ran over
func (d *detector) collectPatternMatch(
	pattern compiledPattern,
	searchText string,
	originalText string,
	m [2]int,
	whitelist map[string]bool,
	isNormalized bool,
	existingIndices map[int]bool,
	results *[]MatchResult,
	mapper *normMapper,
) {
	matchedText := searchText[m[0]:m[1]]
	if d.isWhitelistedMatch(searchText, matchedText, m[0], whitelist) {
		return
	}

	if isNormalized {
		d.collectMappedMatch(pattern, matchedText, m[0], whitelist, existingIndices, results, mapper)
		return
	}

	if existingIndices[m[0]] {
		return
	}
	*results = append(*results, MatchResult{
		Word:     matchedText,
		Root:     pattern.root,
		Index:    m[0],
		Severity: pattern.severity,
		Category: pattern.category,
		Method:   MethodPattern,
	})
	existingIndices[m[0]] = true
}

// isWhitelistedMatch reports whether the matched text or its surrounding word
// (raw or normalized) is whitelisted.
func (d *detector) isWhitelistedMatch(searchText string, matchedText string, matchIndex int, whitelist map[string]bool) bool {
	if whitelist[matchedText] || whitelist[d.normalizeFn(matchedText)] {
		return true
	}
	surrounding := getSurroundingWord(searchText, matchIndex, len(matchedText))
	return whitelist[surrounding] || whitelist[d.normalizeFn(surrounding)]
}

// collectMappedMatch maps a match on normalized text back to the original
// word and records it, dropping matches whose original word is whitelisted
// or is a bare digit token.
func (d *detector) collectMappedMatch(
	pattern compiledPattern,
	matchedText string,
	matchIndex int,
	whitelist map[string]bool,
	existingIndices map[int]bool,
	results *[]MatchResult,
	mapper *normMapper,
) {
	word, origIndex, ok := mapper.get().lookup(matchIndex)
	if !ok || whitelist[strings.ToLower(word)] {
		return
	}
	if isBareDigitTail(word) {
		return
	}
	if existingIndices[origIndex] {
		return
	}
	*results = append(*results, MatchResult{
		Word:     word,
		Root:     pattern.root,
		Index:    origIndex,
		Severity: pattern.severity,
		Category: pattern.category,
		Method:   MethodPattern,
	})
	existingIndices[origIndex] = true
}

// isBareDigitTail reports whether the word ends with digits preceded only by
// non-digits (former endsWithDigitsRe && nonDigitDigitsRe pair). \d is [0-9].
func isBareDigitTail(word string) bool {
	end := len(word)
	for end > 0 && word[end-1] >= '0' && word[end-1] <= '9' {
		end--
	}
	if end == len(word) || end == 0 {
		return false
	}
	for i := range end {
		c := word[i]
		if c >= '0' && c <= '9' {
			return false
		}
	}
	return true
}

// newFuzzyCandidates precomputes per-entry rune counts (both matchers) and
// bigram sets (Dice only) once per detectFuzzy call.
func newFuzzyCandidates(algorithm FuzzyAlgorithm, threshold float64, wordSlice []string) *fuzzyCandidates {
	cand := &fuzzyCandidates{
		matcher:    GetFuzzyMatcher(algorithm),
		algorithm:  algorithm,
		threshold:  threshold,
		runeCounts: make([]int, len(wordSlice)),
	}
	if algorithm == FuzzyDice {
		cand.bigrams = make([]map[string]struct{}, len(wordSlice))
	}
	for i, entry := range wordSlice {
		rc := utf8.RuneCountInString(entry)
		cand.runeCounts[i] = rc
		if algorithm == FuzzyDice && rc >= 2 {
			cand.bigrams[i] = bigrams(entry)
		}
	}
	return cand
}

// fuzzyCandidates precomputes per-entry comparison data once per detectFuzzy
// call and provides provably safe, matcher-specific candidate pruning:
//
//   - Levenshtein: dist >= |m-n|, so sim <= 1-|m-n|/maxLen — a word/entry pair
//     whose bound is below the threshold can never pass, skip it.
//   - Dice: intersection <= min(|A|,|B|), so dice <= 2·min/(|A|+|B|) — bigram
//     counts give the same guarantee. A rune-length bound is NOT valid for Dice.
type fuzzyCandidates struct {
	matcher    FuzzyMatchFn
	algorithm  FuzzyAlgorithm
	threshold  float64
	runeCounts []int
	bigrams    []map[string]struct{} // populated for Dice only
}

func (fc *fuzzyCandidates) skip(wordRC int, wordBG map[string]struct{}, idx int) bool {
	entryRC := fc.runeCounts[idx]
	switch fc.algorithm {
	case FuzzyDice:
		entryBG := fc.bigrams[idx]
		if entryBG == nil || wordBG == nil {
			return false // DiceSimilarity has special cases below 2 runes
		}
		a, b := len(entryBG), len(wordBG)
		bound := 2.0 * float64(min(a, b)) / float64(a+b)
		return bound < fc.threshold
	default: // Levenshtein
		maxLen := max(wordRC, entryRC)
		if maxLen == 0 {
			return false
		}
		diff := wordRC - entryRC
		if diff < 0 {
			diff = -diff
		}
		return 1.0-float64(diff)/float64(maxLen) < fc.threshold
	}
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

	d.mu.RLock()
	wordSlice := d.normalizedWordSlice
	wordToRoot := d.normalizedWordToRoot
	d.mu.RUnlock()

	cand := newFuzzyCandidates(algorithm, threshold, wordSlice)

	existingIndices := buildExistingIndices(*results)

	startTime := time.Now()

	for wi := range origWords {
		if time.Since(startTime).Milliseconds() > regexTimeoutMs {
			break
		}

		word := ""
		if wi < len(normWords) {
			word = normWords[wi]
		}

		wordRC := utf8.RuneCountInString(word)
		if wordRC < 3 || whitelist[word] {
			continue
		}

		byteIndex := 0
		if wi < len(origPositions) {
			byteIndex = origPositions[wi]
		}

		var wordBG map[string]struct{}
		if algorithm == FuzzyDice && wordRC >= 2 {
			wordBG = bigrams(word)
		}

		d.matchFuzzyWord(origWords[wi], word, byteIndex, wordRC, wordBG, wordSlice, wordToRoot, cand, existingIndices, results)
	}
}

// matchFuzzyWord compares word against every dictionary entry of at least
// three runes and records the first hit at or above threshold, using the
// matcher-specific prune to avoid full similarity computations.
func (d *detector) matchFuzzyWord(
	origWord string,
	word string,
	byteIndex int,
	wordRC int,
	wordBG map[string]struct{},
	normDicts []string,
	wordToRoot map[string]string,
	cand *fuzzyCandidates,
	existingIndices map[int]bool,
	results *[]MatchResult,
) {
	for i, normDict := range normDicts {
		if cand.runeCounts[i] < 3 || cand.skip(wordRC, wordBG, i) {
			continue
		}
		if cand.matcher(word, normDict) < cand.threshold {
			continue
		}
		if !existingIndices[byteIndex] {
			entry := d.dict.findRootForWord(wordToRoot[normDict])
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
