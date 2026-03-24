package main

import (
	"math"
	"strings"
	"unicode"

	"golang.org/x/text/unicode/norm"
)

// langConfig holds language-specific metadata (no word data).
type langConfig struct {
	name   string
	locale string
	vowels string
}

var langConfigs = map[string]*langConfig{
	"tr": {name: "Turkce", locale: "tr", vowels: "aeıioöuü"},
	"en": {name: "English", locale: "en", vowels: "aeiou"},
	"es": {name: "Espanol", locale: "es", vowels: "aeiou"},
	"de": {name: "Deutsch", locale: "de", vowels: "aeiouäöü"},
}

// example represents a single generated dataset entry.
type example struct {
	Text       string   `json:"text"`
	Label      int      `json:"label"`
	Root       string   `json:"root"`
	Difficulty string   `json:"difficulty"`
	Transforms []string `json:"transforms"`
	Category   string   `json:"category"`
}

// difficulty weights and transform count ranges.
var difficultyWeights = []struct {
	name   string
	weight float64
}{
	{"easy", 0.25},
	{"medium", 0.35},
	{"hard", 0.25},
	{"extreme", 0.15},
}

var difficultyTransforms = map[string][2]int{
	"easy":    {0, 1},
	"medium":  {1, 2},
	"hard":    {2, 3},
	"extreme": {3, 5},
}

func assignDifficulty(rand func() float64) string {
	r := rand()
	cum := 0.0
	for _, dw := range difficultyWeights {
		cum += dw.weight
		if r <= cum {
			return dw.name
		}
	}
	return "medium"
}

func selectTransforms(difficulty string, rand func() float64) []transform {
	bounds := difficultyTransforms[difficulty]
	min, max := bounds[0], bounds[1]
	count := min + int(math.Floor(rand()*float64(max-min+1)))
	if count == 0 {
		return nil
	}

	// Shuffle transforms using rand
	shuffled := make([]transform, len(transforms))
	copy(shuffled, transforms)
	for i := len(shuffled) - 1; i > 0; i-- {
		j := int(math.Floor(rand() * float64(i+1)))
		shuffled[i], shuffled[j] = shuffled[j], shuffled[i]
	}

	var selected []transform
	familyCounts := make(map[string]int)

	for _, t := range shuffled {
		if len(selected) >= count {
			break
		}
		fc := familyCounts[t.family]
		familyMax := 2
		if t.family == "substitution" && difficulty != "extreme" {
			familyMax = 1
		}
		if fc < familyMax {
			selected = append(selected, t)
			familyCounts[t.family] = fc + 1
		}
	}

	return selected
}

func renderPositiveExample(data *dataSet, lang *langConfig, rand func() float64) example {
	root := data.rootsPositive[int(math.Floor(rand()*float64(len(data.rootsPositive))))]
	difficulty := assignDifficulty(rand)
	selectedTransforms := selectTransforms(difficulty, rand)

	word := root
	var appliedTransforms []string

	for _, t := range selectedTransforms {
		before := word
		word = t.fn(word, data, lang, rand)
		if word != before {
			appliedTransforms = append(appliedTransforms, t.name)
		}
	}

	// Template (70%) vs Context (30%)
	var text string
	useTemplate := rand() < 0.7
	if useTemplate && len(data.templatesPositive) > 0 {
		tpl := data.templatesPositive[int(math.Floor(rand()*float64(len(data.templatesPositive))))]
		text = strings.Replace(tpl, "{word}", word, 1)
	} else if len(data.contextsPositive) > 0 {
		ctx := data.contextsPositive[int(math.Floor(rand()*float64(len(data.contextsPositive))))]
		text = strings.Replace(ctx, "{word}", word, 1)
	} else {
		text = word
	}

	if appliedTransforms == nil {
		appliedTransforms = []string{}
	}

	return example{
		Text:       norm.NFC.String(text),
		Label:      1,
		Root:       root,
		Difficulty: difficulty,
		Transforms: appliedTransforms,
		Category:   "positive",
	}
}

func renderNegativeExample(data *dataSet, rand func() float64) example {
	root := data.rootsNegative[int(math.Floor(rand()*float64(len(data.rootsNegative))))]

	var text string
	useTemplate := rand() < 0.7
	if useTemplate && len(data.templatesNegative) > 0 {
		tpl := data.templatesNegative[int(math.Floor(rand()*float64(len(data.templatesNegative))))]
		text = strings.Replace(tpl, "{word}", root, 1)
	} else if len(data.contextsNegative) > 0 {
		ctx := data.contextsNegative[int(math.Floor(rand()*float64(len(data.contextsNegative))))]
		text = strings.Replace(ctx, "{word}", root, 1)
	} else {
		text = root
	}

	// Minor case variation for some negatives
	if rand() < 0.15 {
		runes := []rune(text)
		if len(runes) > 0 {
			runes[0] = unicode.ToUpper(runes[0])
			text = string(runes)
		}
	}

	return example{
		Text:       norm.NFC.String(text),
		Label:      0,
		Root:       root,
		Difficulty: "clean",
		Transforms: []string{},
		Category:   "negative",
	}
}

// shuffle performs Fisher-Yates shuffle with seeded PRNG.
func shuffle(arr []example, rand func() float64) {
	for i := len(arr) - 1; i > 0; i-- {
		j := int(math.Floor(rand() * float64(i+1)))
		arr[i], arr[j] = arr[j], arr[i]
	}
}
