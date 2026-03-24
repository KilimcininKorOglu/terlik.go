package terlik_test

import (
	"bufio"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"runtime"
	"strings"
	"testing"

	terlik "github.com/KilimcininKorOglu/terlik.go"
)

// spdgEntry represents a single SPDG-generated dataset entry.
type spdgEntry struct {
	Text       string   `json:"text"`
	Label      int      `json:"label"`
	Root       string   `json:"root"`
	Difficulty string   `json:"difficulty"`
	Transforms []string `json:"transforms"`
	Category   string   `json:"category"`
}

// Positive detection rate thresholds by difficulty level.
var positiveThresholds = map[string]*float64{
	"easy":    float64Ptr(85),
	"medium":  float64Ptr(65),
	"hard":    float64Ptr(40),
	"extreme": nil, // report only, no threshold
}

// Maximum false positive rate for negative examples (%).
const falsePositiveLimit = 5.0

func float64Ptr(v float64) *float64 { return &v }

// spdgOutputDir returns the path to SPDG output directory relative to this test file.
func spdgOutputDir() string {
	_, filename, _, _ := runtime.Caller(0)
	return filepath.Join(filepath.Dir(filename), "..", "tools", "spdg", "output")
}

func parseJSONL(filePath string) ([]spdgEntry, error) {
	f, err := os.Open(filePath)
	if err != nil {
		return nil, err
	}
	defer f.Close()

	var entries []spdgEntry
	scanner := bufio.NewScanner(f)
	scanner.Buffer(make([]byte, 1024*1024), 1024*1024)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		var entry spdgEntry
		if err := json.Unmarshal([]byte(line), &entry); err != nil {
			return nil, fmt.Errorf("invalid JSONL line: %w", err)
		}
		entries = append(entries, entry)
	}
	return entries, scanner.Err()
}

func TestSPDG(t *testing.T) {
	languages := []string{"tr", "en", "es", "de"}
	outputDir := spdgOutputDir()

	for _, lang := range languages {
		lang := lang
		t.Run(strings.ToUpper(lang), func(t *testing.T) {
			jsonlPath := filepath.Join(outputDir, fmt.Sprintf("export-%s.jsonl", lang))
			if _, err := os.Stat(jsonlPath); os.IsNotExist(err) {
				t.Skipf("SPDG output not found: %s (run: go run ./tools/spdg --%s --pos 500 --neg 500 --seed 42)", jsonlPath, lang)
				return
			}

			entries, err := parseJSONL(jsonlPath)
			if err != nil {
				t.Fatalf("Failed to parse JSONL: %v", err)
			}

			instance := mustNew(t, &terlik.Options{Language: lang})

			var positives, negatives []spdgEntry
			byDifficulty := make(map[string][]spdgEntry)

			for _, e := range entries {
				if e.Label == 1 {
					positives = append(positives, e)
					byDifficulty[e.Difficulty] = append(byDifficulty[e.Difficulty], e)
				} else {
					negatives = append(negatives, e)
				}
			}

			t.Run("PositiveDetection", func(t *testing.T) {
				for difficulty, group := range byDifficulty {
					detected := 0
					for _, entry := range group {
						if instance.ContainsProfanity(entry.Text, nil) {
							detected++
						}
					}
					rate := float64(detected) / float64(len(group)) * 100
					threshold := positiveThresholds[difficulty]

					if threshold != nil {
						t.Logf("[%s] %s: %d/%d (%.1f%%) — min %.0f%%",
							strings.ToUpper(lang), difficulty, detected, len(group), rate, *threshold)
						if rate < *threshold {
							t.Errorf("[%s] %s detection rate %.1f%% < threshold %.0f%%",
								strings.ToUpper(lang), difficulty, rate, *threshold)
						}
					} else {
						t.Logf("[%s] %s: %d/%d (%.1f%%) — report only",
							strings.ToUpper(lang), difficulty, detected, len(group), rate)
					}
				}
			})

			t.Run("NegativeFalsePositive", func(t *testing.T) {
				if len(negatives) == 0 {
					t.Skip("No negative examples")
					return
				}

				falsePositives := 0
				var fpExamples []string

				for _, entry := range negatives {
					if instance.ContainsProfanity(entry.Text, nil) {
						falsePositives++
						if len(fpExamples) < 10 {
							fpExamples = append(fpExamples, fmt.Sprintf("%q (root: %s)", entry.Text, entry.Root))
						}
					}
				}

				fpRate := float64(falsePositives) / float64(len(negatives)) * 100
				t.Logf("[%s] False positive: %d/%d (%.1f%%)",
					strings.ToUpper(lang), falsePositives, len(negatives), fpRate)

				if len(fpExamples) > 0 {
					t.Logf("[%s] FP examples: %s", strings.ToUpper(lang), strings.Join(fpExamples, ", "))
				}

				if fpRate >= falsePositiveLimit {
					t.Errorf("[%s] False positive rate %.1f%% >= %.0f%%",
						strings.ToUpper(lang), fpRate, falsePositiveLimit)
				}
			})
		})
	}
}
