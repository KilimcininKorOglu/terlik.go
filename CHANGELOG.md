# Changelog

All notable changes to this project will be documented in this file.

## [1.0.1] - 2026-09-20

### Added
- Hot-path benchmarks for detection, cleaning and fuzzy matching
- MIT LICENSE file and a CI workflow (build, vet, race tests, gofmt, golangci-lint, gocyclo, gosec, govulncheck)
- Scheduled workflow that deletes leftover run artifacts and the caches of dead branches and closed pull requests

### Changed
- Detection no longer rescans the text per match and no longer rebuilds the offset mapping for every hit
- Go floor raised to 1.25.13 and `golang.org/x/text` updated to v0.39.0
- Cyclomatic complexity reduced below 10 across the package, with modernized idioms
- SPDG annotates its intentional uint32 wraparound and reports writer errors
- CI generates the SPDG datasets so `TestSPDG` runs instead of skipping
- README states the Go 1.25 requirement
- Source files reformatted with gofmt

### Fixed
- Per-call `MinSeverity` override assertion in the test suite

## [1.0.0] - 2026-03-24

### Added
- Complete Go port of terlik.js profanity detection engine with 4-language support (tr, en, es, de)
- 10-stage text normalization pipeline (invisible chars, NFKD, combining marks, locale lowercase, Cyrillic confusables, char folding, number expansion, leet decode, punctuation removal, repeat collapse)
- 3-pass detection pipeline (locale-lowered, fully normalized, CamelCase decompound)
- Fuzzy matching with Levenshtein distance and Dice coefficient algorithms
- Turkish suffix engine (112 suffixes for agglutinative form detection)
- Thread-safe pattern cache with sync.RWMutex
- Lazy compilation with optional background warmup via goroutine
- Dictionary extension and runtime word management (AddWords/RemoveWords)
- Three detection modes: strict, balanced, loose
- Three mask styles: stars, partial, replace
- Severity and category filtering (MinSeverity, ExcludeCategories)
- Synthetic Profanity Dataset Generator (SPDG) tool with 13 text transforms
- SPDG automated detection tests with difficulty-based thresholds
- Comprehensive README with API reference and architecture documentation

### Changed
- Update golang.org/x/text from v0.21.0 to v0.35.0 for current Unicode tables
- Update Go version directive from 1.22 to 1.25
- Update module path to github.com/KilimcininKorOglu/terlik.go

### Fixed
- Resolve data race in recompile() with sync.RWMutex and proper lock scoping
- Prevent Pass 1 byte offset mismatch for Turkish locale (I/i byte length divergence)
- Fix leading whitespace off-by-one in mapNormalizedToOriginal
- Use call-level timeout instead of per-pattern timeout in runPatterns
- Preserve first occurrence in normalizedWordToRoot to prevent cross-entry collisions
- Use rune count instead of byte length in deduplicateResults comparison
- Filter overlapping matches before text replacement in CleanText
- Allocate fresh slice in removeWords to prevent GC leak
- Skip Pass 1 when locale lowercase changes byte length (Turkish I regression fix)
- Skip invisible-char-only segments in normOffset tracking
- Sort GetSupportedLanguages output for deterministic order
- Add variant-to-root reverse lookup map for O(1) dictionary lookups
- Build normalized lookup outside lock in recompile for reduced lock hold time
- Validate numeric CLI arguments in SPDG and exit on parse error
- Quote all string fields in SPDG CSV writer
- Eliminate duplicate examples in SPDG via deduplication with retry
