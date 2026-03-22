package terlik_test

import (
	"strings"
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
	"time"
)

// ═══════════════════════════════════════════════════
// ENGLISH — adversarial samples
// ═══════════════════════════════════════════════════

func TestENAdversarialAccentedBypass(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	detected := []struct{ input, desc string }{
		{"f\u00FCck", "fück (ü=U+00FC)"},
		{"f\u00F9ck", "fùck (ù=U+00F9)"},
		{"f\u00FBck", "fûck (û=U+00FB)"},
		{"sh\u00EEt", "shît (î=U+00EE)"},
		{"sh\u00EFt", "shït (ï=U+00EF)"},
		{"b\u00EEtch", "bîtch (î=U+00EE)"},
		{"c\u00FCnt", "cünt (ü=U+00FC)"},
		{"d\u00ECck", "dìck (ì=U+00EC)"},
		{"n\u00EFgger", "nïgger (ï=U+00EF)"},
		{"f\u00E0ggot", "fàggot (à=U+00E0)"},
		{"r\u00E0pe", "ràpe (à=U+00E0)"},
		{"p\u00F6rn", "pörn (ö=U+00F6)"},
	}
	for _, tt := range detected {
		t.Run(tt.desc, func(t *testing.T) { assertDetects(t, en, tt.input) })
	}
	t.Run("no FP wönk", func(t *testing.T) { assertClean(t, en, "w\u00F6nk") })
}

func TestENAdversarialCyrillicBypass(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	detected := []struct{ input, desc string }{
		{"fu\u0441k", "fuсk (Cyrillic с)"},
		{"f\u0443ck", "fуck (Cyrillic у)"},
		{"\u0430ss", "аss (Cyrillic а)"},
		{"sh\u0456t", "shіt (Cyrillic і)"},
		{"b\u0456tch", "bіtch (Cyrillic і)"},
		{"\u0441unt", "сunt (Cyrillic с)"},
		{"di\u0441k", "diсk (Cyrillic с)"},
		{"wh\u043Ere", "whоre (Cyrillic о)"},
		{"r\u0430pe", "rаpe (Cyrillic а)"},
		{"p\u043Ern", "pоrn (Cyrillic о)"},
	}
	for _, tt := range detected {
		t.Run(tt.desc, func(t *testing.T) { assertDetects(t, en, tt.input) })
	}
}

func TestENAdversarialFullwidthBypass(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	detected := []struct{ input, desc string }{
		{"\uFF46\uFF55\uFF43\uFF4B", "ｆｕｃｋ (fullwidth)"},
		{"\uFF53\uFF48\uFF49\uFF54", "ｓｈｉｔ (fullwidth)"},
		{"f\uFF55ck", "fｕck (mixed fullwidth u)"},
	}
	for _, tt := range detected {
		t.Run(tt.desc, func(t *testing.T) { assertDetects(t, en, tt.input) })
	}
}

func TestENAdversarialUnicodeNormalization(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	t.Run("NFD combining diacritic fuc+cedilla+k", func(t *testing.T) {
		assertDetects(t, en, "fuc\u0327k")
	})
	t.Run("NFC precomposed fuçk", func(t *testing.T) {
		assertDetects(t, en, "fu\u00E7k")
	})
	t.Run("NFD and NFC consistent for shît", func(t *testing.T) {
		assertDetects(t, en, "shi\u0302t") // NFD
		assertDetects(t, en, "sh\u00EEt")  // NFC
	})
}

func TestENAdversarialZeroWidth(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	t.Run("ZWSP", func(t *testing.T) { assertDetects(t, en, "f\u200Buck") })
	t.Run("ZWNJ", func(t *testing.T) { assertDetects(t, en, "f\u200Cu\u200Cc\u200Ck") })
	t.Run("soft hyphen", func(t *testing.T) { assertDetects(t, en, "f\u00ADu\u00ADc\u00ADk") })
}

func TestENAdversarialFPStress(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	fpTraps := []string{
		"assumption", "cocky", "therapists", "grapevine",
		"passionate", "compassionate", "embarrass", "harassment",
		"scrapbook", "cumulonimbus", "cumulative", "circumvent",
		"pennant", "penalize", "peninsula", "penetrate",
		"Titanic", "constitution", "analytical", "psychoanalysis",
		"masseuse", "cassette", "classic", "classy",
		"Dickensian", "cocktails", "peacocking",
		"buttress", "butterscotch", "rebuttal",
		"sextant", "sextet", "Sussex",
		"shitake", "document", "buckle",
		"Hancock", "cocktail", "shuttlecocks",
	}
	for _, word := range fpTraps {
		t.Run("no FP: "+word, func(t *testing.T) { assertClean(t, en, word) })
	}
}

func TestENAdversarialCompoundEvasion(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	compounds := []string{"fuckwad", "shitlord", "cockwomble", "twatwaffle", "assmunch", "cumguzzler", "dickweasel"}
	for _, word := range compounds {
		t.Run(word, func(t *testing.T) { assertDetects(t, en, word) })
	}
}

func TestENAdversarialBoundaryAttacks(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	t.Run("profanity in URL path", func(t *testing.T) { assertDetects(t, en, "visit example.com/fuck") })
	t.Run("profanity in email", func(t *testing.T) { assertDetects(t, en, "email fuck@email.com") })
	t.Run("CamelCase FuckYou", func(t *testing.T) { assertDetects(t, en, "FuckYou") })
	t.Run("hyphenated mother-fucker", func(t *testing.T) { assertDetects(t, en, "mother-fucker") })
	t.Run("hashtag #fuckyou", func(t *testing.T) { assertDetects(t, en, "#fuckyou") })
}

// ═══════════════════════════════════════════════════
// TURKISH — adversarial samples
// ═══════════════════════════════════════════════════

func TestTRAdversarialLocaleEdgeCases(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("SİKTİR dotted İ capitals", func(t *testing.T) { assertDetects(t, tr, "SİKTİR") })
	t.Run("SIKTIR ASCII capitals", func(t *testing.T) {
		// I→ı (Turkish), charMap ı→i → siktir → detected
		assertDetects(t, tr, "SIKTIR")
	})
}

func TestTRAdversarialAccentedBypass(t *testing.T) {
	tr := mustNew(t, nil)

	detected := []struct{ input, desc string }{
		{"s\u00ECktir", "sìktir (ì=U+00EC)"},
		{"s\u00EEktir", "sîktir (î=U+00EE)"},
		{"or\u00F2spu", "oròspu (ò=U+00F2)"},
	}
	for _, tt := range detected {
		t.Run(tt.desc, func(t *testing.T) { assertDetects(t, tr, tt.input) })
	}
}

func TestTRAdversarialCyrillicBypass(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("Cyrillic і for i", func(t *testing.T) { assertDetects(t, tr, "s\u0456kt\u0456r") })
	t.Run("Cyrillic о for o", func(t *testing.T) { assertDetects(t, tr, "\u043Er\u043Espu") })
}

func TestTRAdversarialFPStress(t *testing.T) {
	tr := mustNew(t, nil)

	trSafe := []string{
		"sıkıntı", "sıkıştı", "sıkma", "sıkı", "sıkıcı",
		"ambalaj", "ameliyat", "ambulans", "amazon",
		"bokser", "boksör", "malzeme", "maliyet",
		"memur", "memorial", "memleket", "memento",
		"piknik", "namuslu", "ahlaki",
		"dolunay", "dolap", "dolar", "dolma",
		"sikke", "siklet", "gotik", "gotham",
		"yarasa",
		"sıkmak", "sıkılmak",
	}
	for _, word := range trSafe {
		t.Run("no FP: "+word, func(t *testing.T) { assertClean(t, tr, word) })
	}
}

func TestTRAdversarialNumberExpansion(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("s2k detected", func(t *testing.T) { assertDetects(t, tr, "s2k") })
	t.Run("s100 safe (not between letters)", func(t *testing.T) { assertClean(t, tr, "s100") })
}

func TestTRAdversarialSuffixBoundary(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("orospu + suffix -lar", func(t *testing.T) { assertDetects(t, tr, "orospular") })
}

// ═══════════════════════════════════════════════════
// SPANISH — adversarial samples
// ═══════════════════════════════════════════════════

func TestESAdversarialAccentedBypass(t *testing.T) {
	es := mustNew(t, &terlik.Options{Language: "es"})

	detected := []struct{ input, desc string }{
		{"m\u00ECerda", "mìerda (ì for i)"},
		{"p\u00FBta", "pûta (û for u)"},
		{"c\u00F2ño", "còño (ò for o)"},
		{"h\u00ECjoputa", "hìjoputa (ì for i)"},
		{"p\u00E8ndejo", "pèndejo (è for e)"},
	}
	for _, tt := range detected {
		t.Run(tt.desc, func(t *testing.T) { assertDetects(t, es, tt.input) })
	}
}

func TestESAdversarialCyrillicBypass(t *testing.T) {
	es := mustNew(t, &terlik.Options{Language: "es"})
	t.Run("Cyrillic а in puta", func(t *testing.T) { assertDetects(t, es, "put\u0430") })
}

func TestESAdversarialFPStress(t *testing.T) {
	es := mustNew(t, &terlik.Options{Language: "es"})

	esSafe := []string{
		"computadora", "disputar", "reputacion", "imputar",
		"pollo", "pollito", "polluelo", "folleto", "follaje",
		"particular", "articulo", "vehicular", "calcular",
		"maricopa", "putamen", "polleria",
	}
	for _, word := range esSafe {
		t.Run("no FP: "+word, func(t *testing.T) { assertClean(t, es, word) })
	}
}

// ═══════════════════════════════════════════════════
// GERMAN — adversarial samples
// ═══════════════════════════════════════════════════

func TestDEAdversarialSzInterchange(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})

	t.Run("Scheisse (ss)", func(t *testing.T) { assertDetects(t, de, "Scheisse") })
	t.Run("SCHEISSE (uppercase)", func(t *testing.T) { assertDetects(t, de, "SCHEISSE") })
	t.Run("SCHEIßE (uppercase with ß)", func(t *testing.T) { assertDetects(t, de, "SCHEIßE") })
}

func TestDEAdversarialAccentedBypass(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})

	detected := []struct{ input, desc string }{
		{"f\u00ECck", "fìck (ì for i)"},
		{"f\u00EEck", "fîck (î for i)"},
		{"H\u00F9re", "Hùre (ù for u)"},
		{"F\u00F2tze", "Fòtze (ò for o)"},
		{"Schl\u00E0mpe", "Schlàmpe (à for a)"},
		{"W\u00ECchser", "Wìchser (ì for i)"},
	}
	for _, tt := range detected {
		t.Run(tt.desc, func(t *testing.T) { assertDetects(t, de, tt.input) })
	}
}

func TestDEAdversarialCyrillicBypass(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})

	t.Run("Cyrillic і in Fick", func(t *testing.T) { assertDetects(t, de, "F\u0456ck") })
	t.Run("Cyrillic А in Arsch", func(t *testing.T) { assertDetects(t, de, "\u0410rsch") })
}

func TestDEAdversarialFPStress(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})

	deSafe := []string{
		"schwanger", "schwangerschaft", "geschichte",
		"ficktion", "arschen", "schwanzen",
		"Gesellschaft", "Wirtschaft", "Wissenschaft",
		"Druckerei", "Druckfehler",
	}
	for _, word := range deSafe {
		t.Run("no FP: "+word, func(t *testing.T) { assertClean(t, de, word) })
	}
}

// ═══════════════════════════════════════════════════
// ReDoS & Performance stress
// ═══════════════════════════════════════════════════

func TestReDoSStress(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})
	tr := mustNew(t, nil)

	// Warm up
	en.ContainsProfanity("warmup", nil)
	tr.ContainsProfanity("warmup", nil)

	t.Run("pathological separator pattern EN", func(t *testing.T) {
		input := strings.Repeat(".", 1000) + "fuck"
		start := time.Now()
		en.ContainsProfanity(input, nil)
		if time.Since(start) > 5*time.Second {
			t.Error("took too long")
		}
	})
	t.Run("alternating separator/letter flood EN", func(t *testing.T) {
		parts := make([]string, 500)
		for i := range parts {
			parts[i] = string(rune('a' + (i % 26)))
		}
		input := strings.Join(parts, ".")
		start := time.Now()
		en.ContainsProfanity(input, nil)
		if time.Since(start) > 5*time.Second {
			t.Error("took too long")
		}
	})
	t.Run("maxLength near-matches EN", func(t *testing.T) {
		input := strings.Repeat("fuc ", 2500)
		start := time.Now()
		en.ContainsProfanity(input, nil)
		if time.Since(start) > 10*time.Second {
			t.Error("took too long")
		}
	})
	t.Run("combining marks flood EN", func(t *testing.T) {
		marks := strings.Repeat("\u0300", 100)
		input := "f" + marks + "u" + marks + "c" + marks + "k"
		start := time.Now()
		en.ContainsProfanity(input, nil)
		if time.Since(start) > 5*time.Second {
			t.Error("took too long")
		}
	})
	t.Run("TR suffix chain flood", func(t *testing.T) {
		input := "sik" + strings.Repeat("tirlerinesinin", 700)
		if len(input) > 10000 {
			input = input[:10000]
		}
		start := time.Now()
		tr.ContainsProfanity(input, nil)
		if time.Since(start) > 10*time.Second {
			t.Error("took too long")
		}
	})
}

// ═══════════════════════════════════════════════════
// CROSS-LANGUAGE ISOLATION
// ═══════════════════════════════════════════════════

func TestCrossLanguageIsolation(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})
	tr := mustNew(t, nil)
	de := mustNew(t, &terlik.Options{Language: "de"})
	es := mustNew(t, &terlik.Options{Language: "es"})

	t.Run("EN does not detect TR", func(t *testing.T) { assertClean(t, en, "siktir git") })
	t.Run("EN does not detect DE", func(t *testing.T) { assertClean(t, en, "du Arschloch") })
	t.Run("TR does not detect EN", func(t *testing.T) { assertClean(t, tr, "what the fuck") })
	t.Run("DE does not detect ES", func(t *testing.T) { assertClean(t, de, "hijo de puta") })
	t.Run("ES does not detect TR", func(t *testing.T) { assertClean(t, es, "orospu çocuğu") })
}

// ═══════════════════════════════════════════════════
// COMPREHENSIVE REDOS ATTACK SURFACE
// ═══════════════════════════════════════════════════

func TestAttackSurfaceSeparatorAbuse(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("single separator s.i.k", func(t *testing.T) { assertDetects(t, tr, "s.i.k") })
	t.Run("mixed separators s_i-k.t.i.r", func(t *testing.T) { assertDetects(t, tr, "s_i-k.t.i.r") })
	t.Run("max 3 separators s...i...k", func(t *testing.T) { assertDetects(t, tr, "s...i...k") })
	t.Run("4+ separators caught via normalizer", func(t *testing.T) { assertDetects(t, tr, "s....i....k") })
	t.Run("tab as separator", func(t *testing.T) { assertDetects(t, tr, "s\ti\tk") })
	t.Run("zero-width chars not separator bypass", func(t *testing.T) { assertDetects(t, tr, "s\u200Di\u200Dk") })
}

func TestAttackSurfaceLeetBypass(t *testing.T) {
	tr := mustNew(t, nil)
	en := mustNew(t, &terlik.Options{Language: "en"})

	t.Run("all-leet $1kt1r lan", func(t *testing.T) { assertDetects(t, tr, "$1kt1r lan") })
	t.Run("mixed leet s1ktir", func(t *testing.T) { assertDetects(t, tr, "s1ktir git") })
	t.Run("@ as a in aptal", func(t *testing.T) { assertDetects(t, tr, "@pt@l") })
	t.Run("8ok bok", func(t *testing.T) { assertDetects(t, tr, "8ok gibi") })
	t.Run("combined leet+separator $...1...k", func(t *testing.T) { assertDetects(t, tr, "$...1...k") })
	t.Run("EN leet f*ck", func(t *testing.T) { assertDetects(t, en, "f*ck you") })
	t.Run("EN leet phuck", func(t *testing.T) { assertDetects(t, en, "phuck") })
}

func TestAttackSurfaceCharRepetition(t *testing.T) {
	tr := mustNew(t, nil)
	en := mustNew(t, &terlik.Options{Language: "en"})

	t.Run("repeated vowels siiiiik", func(t *testing.T) { assertDetects(t, tr, "siiiiik") })
	t.Run("repeated consonants sikkkk", func(t *testing.T) { assertDetects(t, tr, "sikkkk") })
	t.Run("extreme 16 i's", func(t *testing.T) { assertDetects(t, tr, "s"+strings.Repeat("i", 16)+"k") })
	t.Run("repeated leet $$$1kt1r", func(t *testing.T) { assertDetects(t, tr, "$$$1kt1r") })
	t.Run("EN fuuuuck", func(t *testing.T) { assertDetects(t, en, "fuuuuck you") })
}

func TestAttackSurfaceUnicodeTricks(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("Turkish uppercase SiKTiR", func(t *testing.T) { assertDetects(t, tr, "SiKTiR") })
	t.Run("ASCII caps SIKTIR", func(t *testing.T) { assertDetects(t, tr, "SIKTIR") })
	t.Run("mixed case sIkTiR", func(t *testing.T) { assertDetects(t, tr, "sIkTiR") })
	t.Run("fullwidth no crash", func(t *testing.T) { tr.ContainsProfanity("\uFF33\uFF29\uFF2B", nil) })
	t.Run("combining diacritics no crash", func(t *testing.T) { tr.ContainsProfanity("s\u0301i\u0303k\u0300", nil) })
}

func TestAttackSurfaceWhitelistIntegrity(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("sikke whitelisted", func(t *testing.T) { assertClean(t, tr, "sikke") })
	t.Run("amsterdam whitelisted", func(t *testing.T) { assertClean(t, tr, "amsterdam") })
	t.Run("leet whitelist s1kke", func(t *testing.T) { assertClean(t, tr, "s1kke") })
	t.Run("whitelist+suffix sikkeleri", func(t *testing.T) { assertClean(t, tr, "sikkeleri") })
}

func TestAttackSurfaceBoundaryAttacks(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("profanity at start", func(t *testing.T) { assertDetects(t, tr, "siktir git") })
	t.Run("profanity at end", func(t *testing.T) { assertDetects(t, tr, "hadi siktir") })
	t.Run("profanity entire string", func(t *testing.T) { assertDetects(t, tr, "siktir") })
	t.Run("between punctuation", func(t *testing.T) { assertDetects(t, tr, "(siktir)") })
	t.Run("inside quotes", func(t *testing.T) { assertDetects(t, tr, "\"siktir\" dedi") })
	t.Run("trailing numbers safe", func(t *testing.T) { assertClean(t, tr, "siktir123") })
	t.Run("surrounded by emojis", func(t *testing.T) { assertDetects(t, tr, "😀 siktir 😀") })
	t.Run("embedded not matched", func(t *testing.T) { assertClean(t, tr, "mesiktin") })
}

func TestAttackSurfaceMultiMatch(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("multiple profanities", func(t *testing.T) {
		results := tr.GetMatches("siktir git orospu cocugu", nil)
		if len(results) < 2 {
			t.Errorf("expected >=2 matches, got %d", len(results))
		}
	})
	t.Run("same word repeated", func(t *testing.T) {
		parts := make([]string, 20)
		for i := range parts {
			parts[i] = "siktir"
		}
		input := strings.Join(parts, " ")
		start := time.Now()
		results := tr.GetMatches(input, nil)
		if time.Since(start) > 30*time.Second {
			t.Error("took too long")
		}
		if len(results) < 1 {
			t.Error("expected at least 1 match")
		}
	})
	t.Run("different roots in succession", func(t *testing.T) {
		results := tr.GetMatches("sik bok got amk ibne", nil)
		if len(results) < 3 {
			t.Errorf("expected >=3 matches, got %d", len(results))
		}
	})
}

func TestAttackSurfaceInputEdgeCases(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("empty string", func(t *testing.T) { assertClean(t, tr, "") })
	t.Run("whitespace only", func(t *testing.T) { assertClean(t, tr, "   \t\n  ") })
	t.Run("single character", func(t *testing.T) { assertClean(t, tr, "a") })
	t.Run("only numbers", func(t *testing.T) { assertClean(t, tr, "12345678901234567890") })
	t.Run("only special chars", func(t *testing.T) { assertClean(t, tr, "!@#$%^&*()") })
	t.Run("long clean text no FP", func(t *testing.T) {
		assertClean(t, tr, strings.Repeat("bu bir test cumlesdir ", 200))
	})
	t.Run("newlines between chars", func(t *testing.T) { assertDetects(t, tr, "s\ni\nk") })
}

func TestAttackSurfaceSuffixHardening(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("root+suffix no separator", func(t *testing.T) { assertDetects(t, tr, "orospuluk") })
	t.Run("root+2 suffixes", func(t *testing.T) { assertDetects(t, tr, "orospuluklar") })
	t.Run("non-suffixable+suffix safe", func(t *testing.T) { assertClean(t, tr, "ama neden") })
	t.Run("suffix with separator evasion", func(t *testing.T) { assertDetects(t, tr, "s.i.k.t.i.r.l.e.r") })
	t.Run("leet+suffix", func(t *testing.T) { assertDetects(t, tr, "$1kt1rler") })
}

func TestAttackSurfaceDetectionRegression(t *testing.T) {
	tr := mustNew(t, nil)
	en := mustNew(t, &terlik.Options{Language: "en"})

	t.Run("TR plain profanity", func(t *testing.T) { assertDetects(t, tr, "siktir git") })
	t.Run("TR leet $1kt1r", func(t *testing.T) { assertDetects(t, tr, "$1kt1r") })
	t.Run("TR separator s.i.k.t.i.r", func(t *testing.T) { assertDetects(t, tr, "s.i.k.t.i.r") })
	t.Run("TR repeated siiiiiktir", func(t *testing.T) { assertDetects(t, tr, "siiiiiktir") })
	t.Run("TR number s1kt1r", func(t *testing.T) { assertDetects(t, tr, "s1kt1r git") })
	t.Run("TR suffix orospuluk", func(t *testing.T) { assertDetects(t, tr, "orospuluk yapma") })
	t.Run("EN plain profanity", func(t *testing.T) { assertDetects(t, en, "fuck off") })
	t.Run("EN leet f*ck", func(t *testing.T) { assertDetects(t, en, "what the f*ck") })
}
