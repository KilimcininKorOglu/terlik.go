# terlik

Go implementation of the [terlik.js](https://github.com/badursun/terlik.js) multi-language profanity detection and filtering engine. Not a naive blacklist -- a multi-layered normalization and pattern engine that catches what simple string matching misses.

Ships with **Turkish** (flagship, full coverage), **English**, **Spanish**, and **German** built-in. Extend at runtime via `ExtendDictionary` or `AddWords`.

## Features

- Multi-language profanity detection with Turkish-first design
- 10-stage text normalization pipeline (leet speak, separators, char repetition, zero-width chars, Cyrillic confusables)
- Turkish suffix engine (112 suffixes, thousands of detectable forms from 147 roots)
- Three detection modes: strict, balanced, loose (with fuzzy matching)
- Two fuzzy algorithms: Levenshtein distance and Dice coefficient
- ReDoS-safe: Go's RE2 engine guarantees O(n) regex execution
- Thread-safe with concurrent-access-ready pattern cache
- Lazy compilation with optional background warmup via goroutine
- Single dependency: `golang.org/x/text` (NFKD normalization only)

## Install

```bash
go get terlik
```

Requires Go 1.22 or later.

## Quick Start

```go
package main

import (
    "fmt"
    "terlik"
)

func main() {
    // Turkish (default)
    tr, _ := terlik.New(nil)
    fmt.Println(tr.ContainsProfanity("siktir git", nil))  // true
    fmt.Println(tr.Clean("siktir git burdan", nil))        // "****** git burdan"

    // English
    en, _ := terlik.New(&terlik.Options{Language: "en"})
    fmt.Println(en.ContainsProfanity("what the fuck", nil)) // true
    fmt.Println(en.ContainsProfanity("siktir git", nil))    // false (Turkish not loaded)

    // Spanish & German
    es, _ := terlik.New(&terlik.Options{Language: "es"})
    de, _ := terlik.New(&terlik.Options{Language: "de"})
    fmt.Println(es.ContainsProfanity("hijo de puta", nil))  // true
    fmt.Println(de.ContainsProfanity("scheisse", nil))      // true
}
```

## What It Catches

| Evasion technique    | Example                              | Detected as     |
|:---------------------|:-------------------------------------|:----------------|
| Plain text           | `siktir`                             | sik              |
| Turkish I/I          | `SIKTIR`                             | sik              |
| Leet speak           | `$1kt1r`, `@pt@l`                   | sik, aptal       |
| Visual leet (TR)     | `8ok`, `6ot`, `i8ne`, `s2k`         | bok, got, ibne, sik |
| Number words (TR)    | `s2mle` (s+iki+mle)                 | sik (sikimle)    |
| Separators           | `s.i.k.t.i.r`, `s_i_k`              | sik              |
| Char repetition      | `siiiiiktir`, `pu$ttt`               | sik, pust        |
| Mixed punctuation    | `or*spu`, `g0t_v3r3n`               | orospu, got      |
| Suffix forms         | `siktiler`, `orospuluk`, `gotune`    | sik, orospu, got |
| Suffix + evasion     | `s.i.k.t.i.r.l.e.r`, `$1kt1rler`   | sik              |
| Zero-width chars     | `s<ZWSP>i<ZWSP>k<ZWSP>t<ZWSP>i<ZWSP>r` | sik          |
| Cyrillic confusables | Latin-looking Cyrillic substitutions | detected         |
| Phonetic (EN)        | `phuck`, `phucking`                  | fuck             |
| Extended leet (EN)   | `8itch`, `s#it`, `ni66er`            | bitch, shit, nigger |

### False Positive Prevention

Whitelist prevents false positives on legitimate words:

```go
tr, _ := terlik.New(nil)
tr.ContainsProfanity("Amsterdam", nil)  // false
tr.ContainsProfanity("sikke", nil)      // false (Ottoman coin)
tr.ContainsProfanity("ambulans", nil)   // false
tr.ContainsProfanity("siklet", nil)     // false (boxing weight class)
tr.ContainsProfanity("memur", nil)      // false
tr.ContainsProfanity("malzeme", nil)    // false
```

## API Reference

### Creating an Instance

```go
// Default (Turkish)
instance, err := terlik.New(nil)

// With options
instance, err := terlik.New(&terlik.Options{
    Language:   "en",
    Mode:       terlik.ModeBalanced,
    MaskStyle:  terlik.MaskStars,
})
```

### Core Methods

```go
// Check if text contains profanity
instance.ContainsProfanity(text string, opts *terlik.DetectOptions) bool

// Get detailed match results
instance.GetMatches(text string, opts *terlik.DetectOptions) []terlik.MatchResult

// Replace profanity with masked text
instance.Clean(text string, opts *terlik.CleanOptions) string
```

### Runtime Dictionary Mutation

```go
// Add custom words (detected as medium severity)
instance.AddWords([]string{"customSlang", "anotherWord"})

// Remove words that cause false positives in your domain
instance.RemoveWords([]string{"damn"})
```

### Multi-Language Warmup

```go
// Create and warm up instances for multiple languages at once
instances, err := terlik.Warmup([]string{"tr", "en", "es", "de"}, nil)

// Use from the map
trInstance := instances["tr"]
trInstance.Clean("siktir git", nil) // <1ms (already warmed up)
```

### Per-Call Option Overrides

```go
// Override detection behavior per call without changing instance defaults
matches := instance.GetMatches("some text", &terlik.DetectOptions{
    Mode:              terlik.ModeStrict,
    EnableFuzzy:       terlik.BoolPtr(true),
    FuzzyThreshold:    0.85,
    FuzzyAlgorithm:    terlik.FuzzyLevenshtein,
    DisableLeetDecode: terlik.BoolPtr(false),
    MinSeverity:       terlik.SeverityHigh,
    ExcludeCategories: []terlik.Category{terlik.CategoryGeneral},
})

// Override clean behavior per call
cleaned := instance.Clean("some text", &terlik.CleanOptions{
    MaskStyle:   terlik.MaskPartial,
    ReplaceMask: "[CENSORED]",
})
```

### MatchResult

```go
type MatchResult struct {
    Word     string      // Matched text as it appears in the original input
    Root     string      // Dictionary root word
    Index    int         // Byte offset in the original text
    Severity Severity    // "high", "medium", "low"
    Category Category    // "sexual", "insult", "slur", "general"
    Method   MatchMethod // "exact", "pattern", "fuzzy"
}
```

### Using TerlikCore Directly

For advanced use cases where you want to supply your own `LanguageConfig` without the built-in registry:

```go
langConfig, _ := terlik.GetLanguageConfig("tr")
core, err := terlik.NewTerlikCore(langConfig, &terlik.Options{
    Mode: terlik.ModeBalanced,
})
core.ContainsProfanity("test", nil)
```

## Configuration

### Options

| Field              | Type              | Default         | Description                                  |
|:-------------------|:------------------|:----------------|:---------------------------------------------|
| `Language`         | `string`          | `"tr"`          | Language code: tr, en, es, de                |
| `Mode`             | `Mode`            | `"balanced"`    | Detection strictness                         |
| `MaskStyle`        | `MaskStyle`       | `"stars"`       | How to mask matched text                     |
| `ReplaceMask`      | `string`          | `"[***]"`       | Replacement text for `MaskReplace` style     |
| `CustomList`       | `[]string`        | `nil`           | Additional words to detect                   |
| `Whitelist`        | `[]string`        | `nil`           | Words to never flag as profanity             |
| `EnableFuzzy`      | `bool`            | `false`         | Enable fuzzy matching                        |
| `FuzzyThreshold`   | `float64`         | `0.8`           | Similarity threshold (0.0-1.0)               |
| `FuzzyAlgorithm`   | `FuzzyAlgorithm`  | `"levenshtein"` | Fuzzy algorithm: levenshtein or dice         |
| `MaxLength`        | `int`             | `10000`         | Maximum input length in runes                |
| `BackgroundWarmup` | `bool`            | `false`         | Compile patterns in a background goroutine   |
| `ExtendDictionary` | `*DictionaryData` | `nil`           | Merge additional dictionary entries          |
| `DisableLeetDecode`| `bool`            | `false`         | Skip leet speak decoding stage               |
| `DisableCompound`  | `bool`            | `false`         | Skip CamelCase decompounding pass            |
| `MinSeverity`      | `Severity`        | `""`            | Filter results below this severity           |
| `ExcludeCategories`| `[]Category`      | `nil`           | Exclude specific categories from results     |

### Detection Modes

| Mode       | Behavior                                        | Best for                |
|:-----------|:------------------------------------------------|:------------------------|
| `strict`   | Normalize + exact word match only               | Minimum false positives |
| `balanced` | Normalize + regex pattern matching              | General use (default)   |
| `loose`    | Pattern matching + fuzzy similarity matching    | Maximum coverage        |

### Mask Styles

| Style     | Input      | Output     |
|:----------|:-----------|:-----------|
| `stars`   | `siktir`   | `******`   |
| `partial` | `siktir`   | `s****r`   |
| `replace` | `siktir`   | `[***]`    |

## Extending the Dictionary

Add language-specific entries at construction time:

```go
instance, err := terlik.New(&terlik.Options{
    Language: "tr",
    ExtendDictionary: &terlik.DictionaryData{
        Version: 1,
        Entries: []terlik.DictionaryEntry{
            {
                Root:       "customword",
                Variants:   []string{"customvariant"},
                Severity:   "high",
                Category:   "insult",
                Suffixable: true,
            },
        },
        Suffixes:  []string{"ler", "lar"},
        Whitelist: []string{"safeword"},
    },
})
```

## Dictionary Coverage

| Language | Status    | Roots | Suffixes | Whitelist | Effective Forms |
|:---------|:----------|------:|---------:|----------:|:----------------|
| Turkish  | Flagship  |   147 |      112 |        95 | ~10,000+        |
| English  | Full      |   138 |       28 |       106 | ~10,000+        |
| Spanish  | Community |    29 |       13 |        21 | ~500+           |
| German   | Community |    28 |        8 |         6 | ~300+           |

"Effective forms" = roots x normalization variants x suffix combinations x evasion patterns.

## Performance

### Lazy Compilation

Pattern compilation is deferred until the first detection call:

```go
instance, _ := terlik.New(nil)          // Near-instant construction
instance.ContainsProfanity("warmup", nil) // First call compiles patterns
instance.ContainsProfanity("test", nil)   // Subsequent calls are fast
```

### Warmup Strategies

```go
// Option A: Background warmup (recommended for servers)
instance, _ := terlik.New(&terlik.Options{BackgroundWarmup: true})

// Option B: Explicit warmup at startup
instance, _ := terlik.New(nil)
instance.ContainsProfanity("warmup", nil)

// Option C: Multi-language warmup
instances, _ := terlik.Warmup([]string{"tr", "en"}, nil)
```

### Thread Safety

Pattern cache is protected by `sync.RWMutex` at the package level. Multiple goroutines can safely use the same `Terlik` instance for detection and cleaning without external synchronization.

### ReDoS Safety

Go's RE2 regex engine guarantees linear-time execution, eliminating catastrophic backtracking. Additionally, a 250ms timeout safety net protects against edge cases in pattern matching.

---

## For Developers

### How It Works

Ten-stage normalization pipeline, then pattern matching:

```
input
  1. strip invisible chars (ZWSP, ZWNJ, soft hyphen, BOM, etc.)
  2. NFKD decompose (fullwidth -> ASCII, precomposed -> base + combining)
  3. strip combining marks (diacritics U+0300-U+036F)
  4. locale-aware lowercase (Turkish: I->i, I->i)
  5. Cyrillic confusable -> Latin replacement
  6. language-specific char folding (charMap)
  7. number expansion between letters (Turkish: s2k -> sikik)
  8. leet speak decode (leetMap)
  9. punctuation removal between letters (s.i.k -> sik)
 10. collapse 3+ repeated chars to 1 + whitespace trim
  -> 3-pass pattern matching (locale-lowered, normalized, CamelCase decompound)
  -> optional fuzzy matching
  -> whitelist filtering
  -> deduplicated results
```

### Architecture

Two-tier design: `Terlik` resolves language config from the registry and delegates to `TerlikCore`.

```
Terlik (terlik.go)
  |-- resolves LanguageConfig from registry (lang.go)
  |-- delegates to TerlikCore

TerlikCore (core.go)
  |-- dictionary (dictionary.go)     -- word entries, whitelist, suffixes
  |-- detector (detector.go)         -- 3-pass detection pipeline
  |     |-- normalizer (normalizer.go) -- 10-stage text normalization
  |     |-- patterns (patterns.go)     -- charClass regex compilation
  |     |-- fuzzy (fuzzy.go)           -- Levenshtein / Dice similarity
  |-- cleaner (cleaner.go)            -- mask/replace matched text

dictdata/*.json                       -- embedded dictionary files (go:embed)
schema.go                             -- dictionary validation + merge
types.go                              -- all types, constants, enums
```

### Build and Test

```bash
go build ./...                        # Build
go test ./...                         # Run all 1340 tests
go vet ./...                          # Static analysis
go test -run TestFunctionName ./...   # Run a single test
go test -v ./tests/                   # Run external tests with verbose output
```

### Test Organization

Tests are split across two packages:

- **`terlik_test.go`** (root, `package terlik`) -- internal tests for unexported functions like `collapseRepeats`, `expandNumbers`, `turkishToLower`
- **`tests/`** (20 files, `package terlik_test`) -- external tests covering the full public API
- **`tests/helpers_test.go`** -- shared test helpers: `mustNew`, `assertDetects`, `assertClean`, `assertDetectsRoot`

### Go-Specific Design Decisions

**RE2 regex engine constraints** shape several design choices in this port:

- No lookbehind/lookahead: word boundaries use `(?:^|[^WORD_CHAR])(pattern)(?:[^WORD_CHAR]|$)` with a capturing group, extracting the inner match via `FindSubmatchIndex`
- No backreferences: `(.)\1{2,}` is replaced by manual rune scanning in `collapseRepeats`
- Number expansion and punctuation removal use manual rune scanning instead of lookbehind patterns
- Turkish locale lowercase uses a custom `turkishToLower` function since `strings.ToLower` does not handle I/I correctly
- All match indices are byte offsets (Go-idiomatic), not character offsets
- Sorted slices ensure deterministic behavior wherever Go map iteration order matters

### Dictionary JSON Schema

```json
{
  "version": 1,
  "suffixes": ["ler", "lar", "..."],
  "entries": [
    {
      "root": "word",
      "variants": ["variant1", "variant2"],
      "severity": "high",
      "category": "sexual",
      "suffixable": true
    }
  ],
  "whitelist": ["safe_word"]
}
```

Validation: max 150 suffixes (1-10 lowercase Unicode letters each), no duplicate roots or whitelist entries. Severity values: `high`, `medium`, `low`. Category values: `sexual`, `insult`, `slur`, `general`.

---

## License

MIT

## Original Project

This is a Go port of [terlik.js](https://github.com/badursun/terlik.js) -- the original TypeScript/JavaScript implementation by [@badursun](https://github.com/badursun).
