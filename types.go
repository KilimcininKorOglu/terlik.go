package terlik

import "regexp"

// Severity represents the severity level of a profanity entry.
type Severity string

const (
	SeverityHigh   Severity = "high"
	SeverityMedium Severity = "medium"
	SeverityLow    Severity = "low"
)

// SeverityOrder maps severity levels to numeric values for comparison.
var SeverityOrder = map[Severity]int{
	SeverityLow:    0,
	SeverityMedium: 1,
	SeverityHigh:   2,
}

// Category represents the content category for profanity entries.
type Category string

const (
	CategorySexual  Category = "sexual"
	CategoryInsult  Category = "insult"
	CategorySlur    Category = "slur"
	CategoryGeneral Category = "general"
)

// Mode controls the detection strictness.
type Mode string

const (
	ModeStrict   Mode = "strict"
	ModeBalanced Mode = "balanced"
	ModeLoose    Mode = "loose"
)

// MaskStyle determines how matched text is masked.
type MaskStyle string

const (
	MaskStars   MaskStyle = "stars"
	MaskPartial MaskStyle = "partial"
	MaskReplace MaskStyle = "replace"
)

// FuzzyAlgorithm selects the fuzzy matching algorithm.
type FuzzyAlgorithm string

const (
	FuzzyLevenshtein FuzzyAlgorithm = "levenshtein"
	FuzzyDice        FuzzyAlgorithm = "dice"
)

// MatchMethod describes how a match was detected.
type MatchMethod string

const (
	MethodExact   MatchMethod = "exact"
	MethodPattern MatchMethod = "pattern"
	MethodFuzzy   MatchMethod = "fuzzy"
)

// WordEntry represents a single entry in the profanity dictionary.
type WordEntry struct {
	Root       string
	Variants   []string
	Severity   Severity
	Category   string
	Suffixable bool
}

// Options configures a Terlik or TerlikCore instance.
type Options struct {
	Language          string
	Mode              Mode
	MaskStyle         MaskStyle
	CustomList        []string
	Whitelist         []string
	EnableFuzzy       bool
	FuzzyThreshold    float64
	FuzzyAlgorithm    FuzzyAlgorithm
	MaxLength         int
	ReplaceMask       string
	BackgroundWarmup  bool
	ExtendDictionary  *DictionaryData
	DisableLeetDecode bool
	DisableCompound   bool
	MinSeverity       Severity
	ExcludeCategories []Category
}

// DetectOptions provides per-call detection overrides.
// Boolean pointer fields distinguish "not set" (nil) from "explicitly false".
type DetectOptions struct {
	Mode              Mode
	EnableFuzzy       *bool
	FuzzyThreshold    float64
	FuzzyAlgorithm    FuzzyAlgorithm
	DisableLeetDecode *bool
	DisableCompound   *bool
	MinSeverity       Severity
	ExcludeCategories []Category
}

// BoolPtr is a helper to create *bool values for DetectOptions.
func BoolPtr(v bool) *bool { return &v }

// CleanOptions provides per-call clean overrides.
type CleanOptions struct {
	DetectOptions
	MaskStyle   MaskStyle
	ReplaceMask string
}

// MatchResult represents a single profanity match found in the input text.
type MatchResult struct {
	Word     string      `json:"word"`
	Root     string      `json:"root"`
	Index    int         `json:"index"`
	Severity Severity    `json:"severity"`
	Category Category    `json:"category,omitempty"`
	Method   MatchMethod `json:"method"`
}

// compiledPattern is a compiled regex pattern for a dictionary entry.
type compiledPattern struct {
	root     string
	severity Severity
	category Category
	regex    *regexp.Regexp
	variants []string
}

// LanguageConfig holds language-specific configuration.
type LanguageConfig struct {
	Locale           string
	CharMap          map[string]string
	LeetMap          map[string]string
	CharClasses      map[string]string
	NumberExpansions  [][2]string
	Dictionary       DictionaryData
}

// NormalizerConfig configures a language-specific normalizer.
type NormalizerConfig struct {
	Locale          string
	CharMap         map[string]string
	LeetMap         map[string]string
	NumberExpansions [][2]string
}

// DictionaryData represents the raw dictionary data structure from JSON.
type DictionaryData struct {
	Version   int               `json:"version"`
	Suffixes  []string          `json:"suffixes"`
	Entries   []DictionaryEntry `json:"entries"`
	Whitelist []string          `json:"whitelist"`
}

// DictionaryEntry represents a single entry in the dictionary JSON.
type DictionaryEntry struct {
	Root       string   `json:"root"`
	Variants   []string `json:"variants"`
	Severity   string   `json:"severity"`
	Category   string   `json:"category"`
	Suffixable bool     `json:"suffixable"`
}
