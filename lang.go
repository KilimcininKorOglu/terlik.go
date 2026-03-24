package terlik

import (
	"embed"
	"encoding/json"
	"fmt"
	"sort"
	"strings"
)

//go:embed dictdata/tr.json
var trDictJSON []byte

//go:embed dictdata/en.json
var enDictJSON []byte

//go:embed dictdata/es.json
var esDictJSON []byte

//go:embed dictdata/de.json
var deDictJSON []byte

// Ensure embed is used
var _ embed.FS

func mustLoadDict(data []byte, name string) DictionaryData {
	var dict DictionaryData
	if err := json.Unmarshal(data, &dict); err != nil {
		panic(fmt.Sprintf("failed to parse %s dictionary: %v", name, err))
	}
	if err := ValidateDictionary(&dict); err != nil {
		panic(fmt.Sprintf("invalid %s dictionary: %v", name, err))
	}
	return dict
}

var (
	trDict = mustLoadDict(trDictJSON, "tr")
	enDict = mustLoadDict(enDictJSON, "en")
	esDict = mustLoadDict(esDictJSON, "es")
	deDict = mustLoadDict(deDictJSON, "de")
)

// TrConfig is the Turkish language configuration.
var TrConfig = LanguageConfig{
	Locale: "tr",
	CharMap: map[string]string{
		"ç": "c", "Ç": "c", "ğ": "g", "Ğ": "g", "ı": "i", "İ": "i",
		"ö": "o", "Ö": "o", "ş": "s", "Ş": "s", "ü": "u", "Ü": "u",
	},
	LeetMap: map[string]string{
		"0": "o", "1": "i", "2": "i", "3": "e", "4": "a",
		"5": "s", "6": "g", "7": "t", "8": "b", "9": "g",
		"@": "a", "$": "s", "!": "i",
	},
	CharClasses: map[string]string{
		"a": "[a4àáâãäå]",
		"b": "[b8ß]",
		"c": "[cçÇ]",
		"d": "[d]",
		"e": "[e3èéêë]",
		"f": "[f]",
		"g": "[gğĞ69]",
		"h": "[h]",
		"i": "[iıİ12ìíîï]",
		"j": "[j]",
		"k": "[k]",
		"l": "[l1]",
		"m": "[m]",
		"n": "[nñ]",
		"o": "[o0öÖòóôõ]",
		"p": "[p]",
		"q": "[qk]",
		"r": "[r]",
		"s": "[s5şŞß]",
		"t": "[t7]",
		"u": "[uüÜùúûv]",
		"v": "[vu]",
		"w": "[w]",
		"x": "[x]",
		"y": "[y]",
		"z": "[z2]",
	},
	NumberExpansions: [][2]string{
		{"100", "yuz"}, {"50", "elli"}, {"10", "on"}, {"2", "iki"},
	},
	Dictionary: trDict,
}

// EnConfig is the English language configuration.
var EnConfig = LanguageConfig{
	Locale:  "en",
	CharMap: map[string]string{},
	LeetMap: map[string]string{
		"0": "o", "1": "i", "3": "e", "4": "a",
		"5": "s", "6": "g", "7": "t", "8": "b",
		"@": "a", "$": "s", "!": "i", "#": "h",
	},
	CharClasses: map[string]string{
		"a": "[a4]",
		"b": "[b8]",
		"c": "[c]",
		"d": "[d]",
		"e": "[e3]",
		"f": "[fph]",
		"g": "[g96]",
		"h": "[h#]",
		"i": "[i1]",
		"j": "[j]",
		"k": "[k]",
		"l": "[l1]",
		"m": "[m]",
		"n": "[n]",
		"o": "[o0]",
		"p": "[p]",
		"q": "[q]",
		"r": "[r]",
		"s": "[s5]",
		"t": "[t7]",
		"u": "[uv]",
		"v": "[vu]",
		"w": "[w]",
		"x": "[x]",
		"y": "[y]",
		"z": "[z]",
	},
	Dictionary: enDict,
}

// EsConfig is the Spanish language configuration.
var EsConfig = LanguageConfig{
	Locale: "es",
	CharMap: map[string]string{
		"ñ": "n", "Ñ": "n",
		"á": "a", "Á": "a", "é": "e", "É": "e",
		"í": "i", "Í": "i", "ó": "o", "Ó": "o",
		"ú": "u", "Ú": "u",
	},
	LeetMap: map[string]string{
		"0": "o", "1": "i", "3": "e", "4": "a",
		"5": "s", "7": "t",
		"@": "a", "$": "s", "!": "i",
	},
	CharClasses: map[string]string{
		"a": "[a4áÁ]",
		"b": "[b8]",
		"c": "[c]",
		"d": "[d]",
		"e": "[e3éÉ]",
		"f": "[f]",
		"g": "[g9]",
		"h": "[h]",
		"i": "[i1íÍ]",
		"j": "[j]",
		"k": "[k]",
		"l": "[l1]",
		"m": "[m]",
		"n": "[nñÑ]",
		"o": "[o0óÓ]",
		"p": "[p]",
		"q": "[q]",
		"r": "[r]",
		"s": "[s5]",
		"t": "[t7]",
		"u": "[uvúÚ]",
		"v": "[vu]",
		"w": "[w]",
		"x": "[x]",
		"y": "[y]",
		"z": "[z]",
	},
	Dictionary: esDict,
}

// DeConfig is the German language configuration.
var DeConfig = LanguageConfig{
	Locale: "de",
	CharMap: map[string]string{
		"ä": "a", "Ä": "a",
		"ö": "o", "Ö": "o",
		"ü": "u", "Ü": "u",
		"ß": "ss",
	},
	LeetMap: map[string]string{
		"0": "o", "1": "i", "3": "e", "4": "a",
		"5": "s", "7": "t",
		"@": "a", "$": "s", "!": "i",
	},
	CharClasses: map[string]string{
		"a": "[a4äÄ]",
		"b": "[b8]",
		"c": "[c]",
		"d": "[d]",
		"e": "[e3]",
		"f": "[f]",
		"g": "[g9]",
		"h": "[h]",
		"i": "[i1]",
		"j": "[j]",
		"k": "[k]",
		"l": "[l1]",
		"m": "[m]",
		"n": "[n]",
		"o": "[o0öÖ]",
		"p": "[p]",
		"q": "[q]",
		"r": "[r]",
		"s": "[s5ß]",
		"t": "[t7]",
		"u": "[uvüÜ]",
		"v": "[vu]",
		"w": "[w]",
		"x": "[x]",
		"y": "[y]",
		"z": "[z]",
	},
	Dictionary: deDict,
}

const coreDictVersion = 1

var registry = map[string]LanguageConfig{
	"tr": TrConfig,
	"en": EnConfig,
	"es": EsConfig,
	"de": DeConfig,
}

// GetLanguageConfig retrieves the configuration for a supported language.
func GetLanguageConfig(lang string) (LanguageConfig, error) {
	config, ok := registry[lang]
	if !ok {
		available := strings.Join(GetSupportedLanguages(), ", ")
		return LanguageConfig{}, fmt.Errorf("unsupported language: %q. Available languages: %s", lang, available)
	}
	if config.Dictionary.Version < coreDictVersion {
		return LanguageConfig{}, fmt.Errorf(
			"dictionary version %d for language %q is below minimum required version %d",
			config.Dictionary.Version, lang, coreDictVersion,
		)
	}
	return config, nil
}

// GetSupportedLanguages returns all available language codes.
func GetSupportedLanguages() []string {
	keys := make([]string, 0, len(registry))
	for k := range registry {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}
