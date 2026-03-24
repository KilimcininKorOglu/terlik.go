package main

import (
	"fmt"
	"math"
	"sort"
	"strings"
)

func printStats(examples []example, lang string) {
	cfg := langConfigs[lang]
	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
	fmt.Printf("  STATISTICS — %s (%s)\n", cfg.name, lang)
	fmt.Println(strings.Repeat("=", 60))

	labels := map[string]int{"positive": 0, "negative": 0}
	difficulties := map[string]int{}
	transformCounts := map[string]int{}
	rootCounts := map[string]int{}

	for _, ex := range examples {
		labels[ex.Category]++
		difficulties[ex.Difficulty]++
		rootCounts[ex.Root]++
		for _, t := range ex.Transforms {
			transformCounts[t]++
		}
	}

	fmt.Println("\n  Label Distribution:")
	fmt.Printf("    Positive (toxic):  %d\n", labels["positive"])
	fmt.Printf("    Negative (clean):  %d\n", labels["negative"])
	fmt.Printf("    Total:             %d\n", len(examples))

	fmt.Println("\n  Difficulty Distribution:")
	diffKeys := sortedKeys(difficulties)
	for _, diff := range diffKeys {
		count := difficulties[diff]
		pct := float64(count) / float64(len(examples)) * 100
		fmt.Printf("    %-10s %6d  (%.1f%%)\n", diff, count, pct)
	}

	fmt.Println("\n  Transform Usage:")
	sortedTransforms := sortedByValue(transformCounts)
	for _, kv := range sortedTransforms {
		fmt.Printf("    %-14s %6d\n", kv.key, kv.value)
	}

	fmt.Println("\n  Root Word Distribution (top 10):")
	sortedRoots := sortedByValue(rootCounts)
	limit := 10
	if len(sortedRoots) < limit {
		limit = len(sortedRoots)
	}
	for _, kv := range sortedRoots[:limit] {
		fmt.Printf("    %-18s %6d\n", kv.key, kv.value)
	}

	fmt.Println()
	fmt.Println(strings.Repeat("=", 60))
}

func validateExamples(examples []example) int {
	fmt.Println("\n  VALIDATION RESULTS:")
	errors := 0

	// Duplicate check
	texts := make(map[string]bool)
	duplicates := 0
	for _, ex := range examples {
		if texts[ex.Text] {
			duplicates++
		}
		texts[ex.Text] = true
	}
	if duplicates > 0 {
		fmt.Printf("    [WARN] %d duplicate texts found\n", duplicates)
	} else {
		fmt.Println("    [OK] No duplicate texts")
	}

	// Label sanity
	posCount := 0
	negCount := 0
	for _, ex := range examples {
		if ex.Label == 1 {
			posCount++
		} else {
			negCount++
		}
	}
	if posCount == 0 || negCount == 0 {
		fmt.Println("    [ERROR] Single-label dataset")
		errors++
	} else {
		fmt.Printf("    [OK] Both labels present (pos: %d, neg: %d)\n", posCount, negCount)
	}

	// Balance check
	ratio := float64(posCount) / float64(posCount+negCount)
	if ratio < 0.3 || ratio > 0.7 {
		fmt.Printf("    [WARN] Imbalanced dataset (positive ratio: %.1f%%)\n", ratio*100)
	} else {
		fmt.Printf("    [OK] Balanced dataset (positive ratio: %.1f%%)\n", ratio*100)
	}

	// Length check
	emptyTexts := 0
	longTexts := 0
	for _, ex := range examples {
		if strings.TrimSpace(ex.Text) == "" {
			emptyTexts++
		}
		if len(ex.Text) > 500 {
			longTexts++
		}
	}
	if emptyTexts > 0 {
		fmt.Printf("    [ERROR] %d empty texts found\n", emptyTexts)
		errors++
	} else {
		fmt.Println("    [OK] No empty texts")
	}
	if longTexts > 0 {
		fmt.Printf("    [WARN] %d texts longer than 500 chars\n", longTexts)
	}

	return errors
}

func printSamples(examples []example, rand func() float64) {
	fmt.Println("\n  SAMPLE OUTPUTS:")

	var positives, negatives []example
	for _, ex := range examples {
		if ex.Label == 1 {
			positives = append(positives, ex)
		} else {
			negatives = append(negatives, ex)
		}
	}

	fmt.Println("\n  --- Positive (toxic) ---")
	limit := 5
	if len(positives) < limit {
		limit = len(positives)
	}
	for i := 0; i < limit; i++ {
		idx := int(math.Floor(rand() * float64(len(positives))))
		ex := positives[idx]
		fmt.Printf("    [%s] %q (root: %s, transforms: %s)\n",
			ex.Difficulty, ex.Text, ex.Root, strings.Join(ex.Transforms, ","))
	}

	fmt.Println("\n  --- Negative (clean) ---")
	limit = 5
	if len(negatives) < limit {
		limit = len(negatives)
	}
	for i := 0; i < limit; i++ {
		idx := int(math.Floor(rand() * float64(len(negatives))))
		ex := negatives[idx]
		fmt.Printf("    %q (root: %s)\n", ex.Text, ex.Root)
	}
}

type keyValue struct {
	key   string
	value int
}

func sortedKeys(m map[string]int) []string {
	keys := make([]string, 0, len(m))
	for k := range m {
		keys = append(keys, k)
	}
	sort.Strings(keys)
	return keys
}

func sortedByValue(m map[string]int) []keyValue {
	kvs := make([]keyValue, 0, len(m))
	for k, v := range m {
		kvs = append(kvs, keyValue{k, v})
	}
	sort.Slice(kvs, func(i, j int) bool {
		return kvs[i].value > kvs[j].value
	})
	return kvs
}
