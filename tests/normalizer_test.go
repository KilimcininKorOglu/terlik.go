package terlik_test

import (
	"strings"
	"terlik"
	"testing"
)

func TestNormalizeFullPipeline(t *testing.T) {
	t.Run("handles combined transformations", func(t *testing.T) {
		tests := []struct{ input, want string }{
			{"S.İ.K.T.İ.R", "siktir"},
			{"$1k7!r", "siktir"},
			{"SIIIKTIR", "siktir"},
			{"  hello   world  ", "hello world"},
		}
		for _, tt := range tests {
			got := terlik.Normalize(tt.input)
			if got != tt.want {
				t.Errorf("Normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})
	t.Run("handles empty/short input", func(t *testing.T) {
		if got := terlik.Normalize(""); got != "" {
			t.Errorf("Normalize('') = %q, want ''", got)
		}
		if got := terlik.Normalize("a"); got != "a" {
			t.Errorf("Normalize('a') = %q, want 'a'", got)
		}
	})
	t.Run("preserves emojis", func(t *testing.T) {
		result := terlik.Normalize("hello 😀 world")
		if !strings.Contains(result, "😀") {
			t.Errorf("expected emoji preserved, got %q", result)
		}
	})
}

func TestNormalizerTurkishLowercase(t *testing.T) {
	// With charMap (ı→i) like real Turkish config — İ→NFKD→I+dot→strip→I→tr_lower→ı→charMap→i
	normalize := terlik.CreateNormalizer(terlik.NormalizerConfig{
		Locale: "tr",
		CharMap: map[string]string{
			"ç": "c", "Ç": "c", "ğ": "g", "Ğ": "g", "ı": "i", "İ": "i",
			"ö": "o", "Ö": "o", "ş": "s", "Ş": "s", "ü": "u", "Ü": "u",
		},
		LeetMap: map[string]string{},
	})
	t.Run("converts to lowercase with Turkish locale", func(t *testing.T) {
		if got := normalize("HELLO"); got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
		got := normalize("İSTANBUL")
		if got != "istanbul" {
			t.Errorf("got %q, want %q", got, "istanbul")
		}
	})
}

func TestNormalizerTurkishChars(t *testing.T) {
	normalize := terlik.CreateNormalizer(terlik.NormalizerConfig{
		Locale: "tr",
		CharMap: map[string]string{
			"ç": "c", "Ç": "c", "ğ": "g", "Ğ": "g", "ı": "i", "İ": "i",
			"ö": "o", "Ö": "o", "ş": "s", "Ş": "s", "ü": "u", "Ü": "u",
		},
		LeetMap: map[string]string{},
	})
	// After NFKD + combining mark strip + locale lower + charMap
	t.Run("replaces Turkish special characters", func(t *testing.T) {
		// "çğıöşü" → NFKD may decompose some, then charMap folds
		got := normalize("çğıöşü")
		if got != "cgiosu" {
			t.Errorf("got %q, want %q", got, "cgiosu")
		}
	})
	t.Run("preserves non-Turkish chars", func(t *testing.T) {
		got := normalize("hello")
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})
}

func TestNormalizerLeetspeak(t *testing.T) {
	normalize := terlik.CreateNormalizer(terlik.NormalizerConfig{
		Locale:  "en",
		CharMap: map[string]string{},
		LeetMap: map[string]string{
			"0": "o", "1": "i", "2": "i", "3": "e", "4": "a",
			"5": "s", "6": "g", "7": "t", "8": "b", "9": "g",
			"@": "a", "$": "s", "!": "i",
		},
	})
	t.Run("replaces common leet substitutions", func(t *testing.T) {
		if got := normalize("h3ll0"); got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
		if got := normalize("$1k"); got != "sik" {
			t.Errorf("got %q, want %q", got, "sik")
		}
		if got := normalize("4m1n4"); got != "amina" {
			t.Errorf("got %q, want %q", got, "amina")
		}
	})
}

func TestNormalizerNumberExpansion(t *testing.T) {
	normalize := terlik.CreateNormalizer(terlik.NormalizerConfig{
		Locale:  "tr",
		CharMap: map[string]string{"ı": "i", "İ": "i"},
		LeetMap: map[string]string{},
		NumberExpansions: [][2]string{
			{"100", "yuz"}, {"50", "elli"}, {"10", "on"}, {"2", "iki"},
		},
	})
	t.Run("expands numbers between letters", func(t *testing.T) {
		if got := normalize("s2k"); got != "sikik" {
			t.Errorf("got %q, want %q", got, "sikik")
		}
		if got := normalize("a2b"); got != "aikib" {
			t.Errorf("got %q, want %q", got, "aikib")
		}
	})
	t.Run("does not expand standalone numbers", func(t *testing.T) {
		if got := normalize("2023 yilinda"); strings.Contains(got, "iki") {
			t.Errorf("standalone number should not be expanded: %q", got)
		}
	})
}

func TestNormalizerPunctuation(t *testing.T) {
	normalize := terlik.CreateNormalizer(terlik.NormalizerConfig{
		Locale:  "en",
		CharMap: map[string]string{},
		LeetMap: map[string]string{},
	})
	t.Run("removes punctuation between letters", func(t *testing.T) {
		tests := []struct{ input, want string }{
			{"s.i.k", "sik"},
			{"s-i-k", "sik"},
			{"s_i_k", "sik"},
			{"s*i*k", "sik"},
		}
		for _, tt := range tests {
			got := normalize(tt.input)
			if got != tt.want {
				t.Errorf("normalize(%q) = %q, want %q", tt.input, got, tt.want)
			}
		}
	})
	t.Run("preserves punctuation at boundaries", func(t *testing.T) {
		got := normalize("test.")
		if got != "test." {
			t.Errorf("got %q, want %q", got, "test.")
		}
	})
}

func TestNormalizerCollapseRepeats(t *testing.T) {
	normalize := terlik.CreateNormalizer(terlik.NormalizerConfig{
		Locale:  "en",
		CharMap: map[string]string{},
		LeetMap: map[string]string{},
	})
	t.Run("collapses 3+ repeated chars to 1", func(t *testing.T) {
		if got := normalize("siiik"); got != "sik" {
			t.Errorf("got %q, want %q", got, "sik")
		}
		if got := normalize("ammmk"); got != "amk" {
			t.Errorf("got %q, want %q", got, "amk")
		}
		if got := normalize("aaaaaa"); got != "a" {
			t.Errorf("got %q, want %q", got, "a")
		}
	})
	t.Run("preserves 2 repeated chars", func(t *testing.T) {
		if got := normalize("oo"); got != "oo" {
			t.Errorf("got %q, want %q", got, "oo")
		}
	})
}

func TestNormalizerTrimWhitespace(t *testing.T) {
	normalize := terlik.CreateNormalizer(terlik.NormalizerConfig{
		Locale:  "en",
		CharMap: map[string]string{},
		LeetMap: map[string]string{},
	})
	t.Run("collapses multiple spaces", func(t *testing.T) {
		if got := normalize("  hello   world  "); got != "hello world" {
			t.Errorf("got %q, want %q", got, "hello world")
		}
	})
}

func TestNormalizerInvisibleChars(t *testing.T) {
	normalize := terlik.CreateNormalizer(terlik.NormalizerConfig{
		Locale:  "en",
		CharMap: map[string]string{},
		LeetMap: map[string]string{},
	})
	t.Run("strips zero-width characters", func(t *testing.T) {
		input := "h\u200Be\u200Cl\u200Dl\u200Eo"
		got := normalize(input)
		if got != "hello" {
			t.Errorf("got %q, want %q", got, "hello")
		}
	})
}
