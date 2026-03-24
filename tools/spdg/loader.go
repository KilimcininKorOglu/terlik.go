package main

import (
	"embed"
	"fmt"
	"io/fs"
	"path"
	"regexp"
	"strconv"
	"strings"
)

//go:embed data
var dataFS embed.FS

// dataSet holds all loaded data for a specific language.
type dataSet struct {
	rootsPositive     []string
	rootsNegative     []string
	templatesPositive []string
	templatesNegative []string
	contextsPositive  []string
	contextsNegative  []string
	suffixes          []string
	leetMap           map[string][]string
	emojiReplacements []string

	separators  []string
	unicodeMap  map[string][]string
	zalgoChars  []string
	zwcChars    []string
}

var unicodeEscapeRe = regexp.MustCompile(`\\u([0-9a-fA-F]{4})`)

func parseUnicodeEscapes(s string) string {
	return unicodeEscapeRe.ReplaceAllStringFunc(s, func(match string) string {
		hex := match[2:]
		code, err := strconv.ParseInt(hex, 16, 32)
		if err != nil {
			return match
		}
		return string(rune(code))
	})
}

func loadTextFile(fsPath string) []string {
	data, err := fs.ReadFile(dataFS, fsPath)
	if err != nil {
		return nil
	}
	lines := strings.Split(string(data), "\n")
	var result []string
	for _, l := range lines {
		l = strings.TrimSpace(l)
		if l != "" && !strings.HasPrefix(l, "#") {
			result = append(result, l)
		}
	}
	return result
}

func loadMapFile(fsPath string) map[string][]string {
	lines := loadTextFile(fsPath)
	m := make(map[string][]string)
	for _, line := range lines {
		parts := strings.SplitN(line, "->", 2)
		if len(parts) != 2 {
			continue
		}
		key := strings.TrimSpace(parts[0])
		vals := strings.Split(parts[1], ",")
		var parsed []string
		for _, v := range vals {
			v = strings.TrimSpace(v)
			if v != "" {
				parsed = append(parsed, parseUnicodeEscapes(v))
			}
		}
		if len(parsed) > 0 {
			m[key] = parsed
		}
	}
	return m
}

func loadUnicodeListFile(fsPath string) []string {
	lines := loadTextFile(fsPath)
	var result []string
	for _, l := range lines {
		result = append(result, parseUnicodeEscapes(l))
	}
	return result
}

func loadAllData(lang string) (*dataSet, error) {
	langDir := path.Join("data", lang)
	sharedDir := path.Join("data", "shared")

	d := &dataSet{
		rootsPositive:     loadTextFile(path.Join(langDir, "roots_positive.txt")),
		rootsNegative:     loadTextFile(path.Join(langDir, "roots_negative.txt")),
		templatesPositive: loadTextFile(path.Join(langDir, "templates_positive.txt")),
		templatesNegative: loadTextFile(path.Join(langDir, "templates_negative.txt")),
		contextsPositive:  loadTextFile(path.Join(langDir, "contexts_positive.txt")),
		contextsNegative:  loadTextFile(path.Join(langDir, "contexts_negative.txt")),
		suffixes:          loadTextFile(path.Join(langDir, "suffixes.txt")),
		leetMap:           loadMapFile(path.Join(langDir, "leet_map.txt")),
		emojiReplacements: loadTextFile(path.Join(langDir, "emoji_replacements.txt")),

		separators: loadTextFile(path.Join(sharedDir, "separators.txt")),
		unicodeMap: loadMapFile(path.Join(sharedDir, "unicode_map.txt")),
		zalgoChars: loadUnicodeListFile(path.Join(sharedDir, "zalgo_chars.txt")),
		zwcChars:   loadUnicodeListFile(path.Join(sharedDir, "zwc_chars.txt")),
	}

	var missing []string
	if len(d.rootsPositive) == 0 {
		missing = append(missing, "roots_positive.txt")
	}
	if len(d.rootsNegative) == 0 {
		missing = append(missing, "roots_negative.txt")
	}
	if len(d.templatesPositive) == 0 {
		missing = append(missing, "templates_positive.txt")
	}
	if len(d.templatesNegative) == 0 {
		missing = append(missing, "templates_negative.txt")
	}

	if len(missing) > 0 {
		return nil, fmt.Errorf("required files empty or missing [%s]: %s", lang, strings.Join(missing, ", "))
	}

	return d, nil
}
