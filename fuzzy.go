package terlik

// LevenshteinDistance computes the edit distance between two strings.
// Uses O(n) space optimization with two-row approach.
func LevenshteinDistance(a, b string) int {
	aRunes := []rune(a)
	bRunes := []rune(b)
	m := len(aRunes)
	n := len(bRunes)

	if m == 0 {
		return n
	}
	if n == 0 {
		return m
	}

	prev := make([]int, n+1)
	curr := make([]int, n+1)

	for j := 0; j <= n; j++ {
		prev[j] = j
	}

	for i := 1; i <= m; i++ {
		curr[0] = i
		for j := 1; j <= n; j++ {
			cost := 1
			if aRunes[i-1] == bRunes[j-1] {
				cost = 0
			}
			del := prev[j] + 1
			ins := curr[j-1] + 1
			sub := prev[j-1] + cost
			curr[j] = min(del, min(ins, sub))
		}
		prev, curr = curr, prev
	}

	return prev[n]
}

// LevenshteinSimilarity computes the similarity ratio (0-1) between two strings.
func LevenshteinSimilarity(a, b string) float64 {
	aLen := len([]rune(a))
	bLen := len([]rune(b))
	maxLen := max(aLen, bLen)
	if maxLen == 0 {
		return 1.0
	}
	return 1.0 - float64(LevenshteinDistance(a, b))/float64(maxLen)
}

func bigrams(s string) map[string]struct{} {
	runes := []rune(s)
	set := make(map[string]struct{})
	for i := 0; i < len(runes)-1; i++ {
		set[string(runes[i:i+2])] = struct{}{}
	}
	return set
}

// DiceSimilarity computes the Dice coefficient (bigram similarity) between two strings.
func DiceSimilarity(a, b string) float64 {
	if len([]rune(a)) < 2 || len([]rune(b)) < 2 {
		if a == b {
			return 1.0
		}
		return 0.0
	}

	bigramsA := bigrams(a)
	bigramsB := bigrams(b)

	intersection := 0
	for bg := range bigramsA {
		if _, ok := bigramsB[bg]; ok {
			intersection++
		}
	}

	return (2.0 * float64(intersection)) / float64(len(bigramsA)+len(bigramsB))
}

// FuzzyMatchFn computes similarity between two strings, returning 0-1.
type FuzzyMatchFn func(a, b string) float64

// GetFuzzyMatcher returns the appropriate fuzzy matching function.
func GetFuzzyMatcher(algorithm FuzzyAlgorithm) FuzzyMatchFn {
	if algorithm == FuzzyDice {
		return DiceSimilarity
	}
	return LevenshteinSimilarity
}
