package terlik

import "fmt"

// Terlik is a multi-language profanity detection and filtering engine.
// It resolves language config from the built-in registry.
type Terlik struct {
	*TerlikCore
}

// New creates a new Terlik instance with the given options.
// Defaults to Turkish ("tr") if no language is specified.
func New(opts *Options) (*Terlik, error) {
	lang := "tr"
	if opts != nil && opts.Language != "" {
		lang = opts.Language
	}

	langConfig, err := GetLanguageConfig(lang)
	if err != nil {
		return nil, fmt.Errorf("failed to get language config: %w", err)
	}

	core, err := NewTerlikCore(langConfig, opts)
	if err != nil {
		return nil, err
	}

	return &Terlik{TerlikCore: core}, nil
}

// Warmup creates and JIT-warms instances for multiple languages at once.
// Returns a map of language code to warmed-up Terlik instance.
func Warmup(languages []string, baseOpts *Options) (map[string]*Terlik, error) {
	if languages == nil {
		languages = GetSupportedLanguages()
	}

	result := make(map[string]*Terlik, len(languages))
	for _, lang := range languages {
		opts := &Options{}
		if baseOpts != nil {
			*opts = *baseOpts
		}
		opts.Language = lang

		instance, err := New(opts)
		if err != nil {
			return nil, fmt.Errorf("failed to create instance for language %q: %w", lang, err)
		}
		instance.ContainsProfanity("warmup", nil)
		result[lang] = instance
	}

	return result, nil
}
