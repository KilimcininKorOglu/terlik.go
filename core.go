package terlik

import (
	"fmt"
	"regexp"
)

// TerlikCore is the core profanity detection and filtering engine.
// Requires a pre-resolved LanguageConfig — no registry dependency.
type TerlikCore struct {
	dict              *dictionary
	det               *detector
	mode              Mode
	maskStyle         MaskStyle
	enableFuzzy       bool
	fuzzyThreshold    float64
	fuzzyAlgorithm    FuzzyAlgorithm
	maxLength         int
	replaceMask       string
	disableLeetDecode bool
	disableCompound   bool
	minSeverity       Severity
	excludeCategories []Category
	language          string
}

// NewTerlikCore creates a new TerlikCore instance with a pre-resolved language config.
func NewTerlikCore(langConfig LanguageConfig, opts *Options) (*TerlikCore, error) {
	c := &TerlikCore{
		language:       langConfig.Locale,
		mode:           ModeBalanced,
		maskStyle:      MaskStars,
		fuzzyAlgorithm: FuzzyLevenshtein,
		fuzzyThreshold: 0.8,
		maxLength:      MaxInputLength,
		replaceMask:    "[***]",
	}

	if opts != nil {
		if opts.Mode != "" {
			c.mode = opts.Mode
		}
		if opts.MaskStyle != "" {
			c.maskStyle = opts.MaskStyle
		}
		if opts.FuzzyAlgorithm != "" {
			c.fuzzyAlgorithm = opts.FuzzyAlgorithm
		}
		if opts.ReplaceMask != "" {
			c.replaceMask = opts.ReplaceMask
		}
		c.enableFuzzy = opts.EnableFuzzy
		c.disableLeetDecode = opts.DisableLeetDecode
		c.disableCompound = opts.DisableCompound
		c.minSeverity = opts.MinSeverity
		c.excludeCategories = opts.ExcludeCategories

		if opts.FuzzyThreshold != 0 {
			if opts.FuzzyThreshold < 0 || opts.FuzzyThreshold > 1 {
				return nil, fmt.Errorf("fuzzyThreshold must be between 0 and 1, got %f", opts.FuzzyThreshold)
			}
			c.fuzzyThreshold = opts.FuzzyThreshold
		}

		if opts.MaxLength != 0 {
			if opts.MaxLength < 1 {
				return nil, fmt.Errorf("maxLength must be at least 1, got %d", opts.MaxLength)
			}
			c.maxLength = opts.MaxLength
		}
	}

	normalizeFn := CreateNormalizer(NormalizerConfig{
		Locale:          langConfig.Locale,
		CharMap:         langConfig.CharMap,
		LeetMap:         langConfig.LeetMap,
		NumberExpansions: langConfig.NumberExpansions,
	})
	safeNormalizeFn := CreateNormalizer(NormalizerConfig{
		Locale:  langConfig.Locale,
		CharMap: langConfig.CharMap,
		LeetMap: map[string]string{},
	})

	dictData := langConfig.Dictionary
	if opts != nil && opts.ExtendDictionary != nil {
		if err := ValidateDictionary(opts.ExtendDictionary); err != nil {
			return nil, fmt.Errorf("extendDictionary validation failed: %w", err)
		}
		dictData = MergeDictionaries(dictData, *opts.ExtendDictionary)
	}

	var customList, whitelist []string
	if opts != nil {
		customList = opts.CustomList
		whitelist = opts.Whitelist
	}

	c.dict = newDictionary(dictData, customList, whitelist)

	hasCustomDict := opts != nil && (len(opts.CustomList) > 0 || len(opts.Whitelist) > 0 || opts.ExtendDictionary != nil)
	cacheKey := langConfig.Locale
	if hasCustomDict {
		cacheKey = ""
	}

	c.det = newDetector(
		c.dict,
		normalizeFn,
		safeNormalizeFn,
		langConfig.Locale,
		langConfig.CharClasses,
		cacheKey,
	)

	if opts != nil && opts.BackgroundWarmup {
		go func() {
			c.det.compile()
			c.ContainsProfanity("warmup", nil)
		}()
	}

	return c, nil
}

// Language returns the language code this instance was created with.
func (c *TerlikCore) Language() string {
	return c.language
}

// ContainsProfanity checks if the text contains profanity.
func (c *TerlikCore) ContainsProfanity(text string, opts *DetectOptions) bool {
	input := validateInput(text, c.maxLength)
	if len(input) == 0 {
		return false
	}
	matches := c.det.detect(input, c.mergeDetectOptions(opts))
	return len(matches) > 0
}

// GetMatches returns all profanity matches found in the text.
func (c *TerlikCore) GetMatches(text string, opts *DetectOptions) []MatchResult {
	input := validateInput(text, c.maxLength)
	if len(input) == 0 {
		return nil
	}
	return c.det.detect(input, c.mergeDetectOptions(opts))
}

// Clean replaces matched profanity with masked versions.
func (c *TerlikCore) Clean(text string, opts *CleanOptions) string {
	input := validateInput(text, c.maxLength)
	if len(input) == 0 {
		return input
	}

	var detectOpts *DetectOptions
	style := c.maskStyle
	replaceMask := c.replaceMask

	if opts != nil {
		detectOpts = &opts.DetectOptions
		if opts.MaskStyle != "" {
			style = opts.MaskStyle
		}
		if opts.ReplaceMask != "" {
			replaceMask = opts.ReplaceMask
		}
	}

	matches := c.det.detect(input, c.mergeDetectOptions(detectOpts))
	return CleanText(input, matches, style, replaceMask)
}

// AddWords adds words to the dictionary at runtime and recompiles patterns.
func (c *TerlikCore) AddWords(words []string) {
	c.dict.addWords(words)
	c.det.recompile()
}

// RemoveWords removes words from the dictionary at runtime and recompiles patterns.
func (c *TerlikCore) RemoveWords(words []string) {
	c.dict.removeWords(words)
	c.det.recompile()
}

// GetPatterns returns the compiled regex patterns keyed by root word.
func (c *TerlikCore) GetPatterns() map[string]*regexp.Regexp {
	return c.det.getPatterns()
}

func (c *TerlikCore) mergeDetectOptions(opts *DetectOptions) *DetectOptions {
	merged := &DetectOptions{
		Mode:              c.mode,
		EnableFuzzy:       BoolPtr(c.enableFuzzy),
		FuzzyThreshold:    c.fuzzyThreshold,
		FuzzyAlgorithm:    c.fuzzyAlgorithm,
		DisableLeetDecode: BoolPtr(c.disableLeetDecode),
		DisableCompound:   BoolPtr(c.disableCompound),
		MinSeverity:       c.minSeverity,
		ExcludeCategories: c.excludeCategories,
	}

	if opts == nil {
		return merged
	}

	if opts.Mode != "" {
		merged.Mode = opts.Mode
	}
	if opts.EnableFuzzy != nil {
		merged.EnableFuzzy = opts.EnableFuzzy
	}
	if opts.FuzzyThreshold != 0 {
		merged.FuzzyThreshold = opts.FuzzyThreshold
	}
	if opts.FuzzyAlgorithm != "" {
		merged.FuzzyAlgorithm = opts.FuzzyAlgorithm
	}
	if opts.DisableLeetDecode != nil {
		merged.DisableLeetDecode = opts.DisableLeetDecode
	}
	if opts.DisableCompound != nil {
		merged.DisableCompound = opts.DisableCompound
	}
	if opts.MinSeverity != "" {
		merged.MinSeverity = opts.MinSeverity
	}
	if opts.ExcludeCategories != nil {
		merged.ExcludeCategories = opts.ExcludeCategories
	}

	return merged
}
