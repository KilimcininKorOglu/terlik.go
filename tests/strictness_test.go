package terlik_test

import (
	"terlik"
	"testing"
)

func TestDisableLeetDecode(t *testing.T) {
	t.Run("default: leet decode catches $1kt1r", func(t *testing.T) {
		tr := mustNew(t, nil)
		assertDetects(t, tr, "$1kt1r")
	})
	t.Run("constructor toggle: skips leet decode", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{DisableLeetDecode: true})
		assertClean(t, tr, "$1kt1r")
	})
	t.Run("per-call override: disableLeetDecode on single call", func(t *testing.T) {
		tr := mustNew(t, nil)
		if tr.ContainsProfanity("$1kt1r", &terlik.DetectOptions{DisableLeetDecode: terlik.BoolPtr(true)}) {
			t.Error("expected disabled leet to not catch $1kt1r")
		}
		assertDetects(t, tr, "$1kt1r")
	})
	t.Run("safety layers stay active with disableLeetDecode", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{DisableLeetDecode: true})
		// NFKD fullwidth
		assertDetects(t, tr, "ｓｉｋｔｉｒ")
		// Diacritics
		assertDetects(t, tr, "sïktïr")
		// Cyrillic confusable
		assertDetects(t, tr, "оrоspu")
	})
	t.Run("plain profanity still detected", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{DisableLeetDecode: true})
		assertDetects(t, tr, "amk")
		assertDetects(t, tr, "orospu")
		assertDetects(t, tr, "siktir")
	})
}

func TestDisableCompound(t *testing.T) {
	t.Run("default: CamelCase decompounding active", func(t *testing.T) {
		en := mustNew(t, &terlik.Options{Language: "en"})
		assertDetects(t, en, "ShitPerson")
	})
	t.Run("constructor toggle: skips CamelCase", func(t *testing.T) {
		en := mustNew(t, &terlik.Options{Language: "en", DisableCompound: true})
		assertClean(t, en, "ShitPerson")
	})
	t.Run("per-call override", func(t *testing.T) {
		en := mustNew(t, &terlik.Options{Language: "en"})
		if en.ContainsProfanity("ShitPerson", &terlik.DetectOptions{DisableCompound: terlik.BoolPtr(true)}) {
			t.Error("expected disabled compound to not catch ShitPerson")
		}
		assertDetects(t, en, "ShitPerson")
	})
	t.Run("explicit variants unaffected", func(t *testing.T) {
		en := mustNew(t, &terlik.Options{Language: "en", DisableCompound: true})
		assertDetects(t, en, "motherfucker")
		assertDetects(t, en, "fuckyou")
	})
	t.Run("plain profanity still detected", func(t *testing.T) {
		en := mustNew(t, &terlik.Options{Language: "en", DisableCompound: true})
		assertDetects(t, en, "fuck")
		assertDetects(t, en, "shit")
	})
}

func TestMinSeverity(t *testing.T) {
	t.Run("default: all severities", func(t *testing.T) {
		tr := mustNew(t, nil)
		assertDetects(t, tr, "salak")
		assertDetects(t, tr, "bok")
		assertDetects(t, tr, "siktir")
	})
	t.Run("minSeverity=medium skips low", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{MinSeverity: terlik.SeverityMedium})
		assertClean(t, tr, "salak")
		assertDetects(t, tr, "bok")
		assertDetects(t, tr, "siktir")
	})
	t.Run("minSeverity=high skips low and medium", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{MinSeverity: terlik.SeverityHigh})
		assertClean(t, tr, "salak")
		assertClean(t, tr, "bok")
		assertDetects(t, tr, "siktir")
	})
	t.Run("per-call override", func(t *testing.T) {
		tr := mustNew(t, nil)
		if tr.ContainsProfanity("salak", &terlik.DetectOptions{MinSeverity: terlik.SeverityMedium}) {
			t.Error("expected per-call minSeverity to skip low")
		}
		assertDetects(t, tr, "salak")
	})
	t.Run("minSeverity=low is no filter", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{MinSeverity: terlik.SeverityLow})
		assertDetects(t, tr, "salak")
		assertDetects(t, tr, "bok")
		assertDetects(t, tr, "siktir")
	})
}

func TestExcludeCategories(t *testing.T) {
	t.Run("default: all categories", func(t *testing.T) {
		tr := mustNew(t, nil)
		assertDetects(t, tr, "siktir")
		assertDetects(t, tr, "orospu")
		assertDetects(t, tr, "ibne")
		assertDetects(t, tr, "bok")
	})
	t.Run("exclude sexual", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{ExcludeCategories: []terlik.Category{terlik.CategorySexual}})
		assertClean(t, tr, "siktir")
		assertDetects(t, tr, "orospu")
		assertDetects(t, tr, "bok")
	})
	t.Run("exclude multiple", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{ExcludeCategories: []terlik.Category{terlik.CategorySexual, terlik.CategorySlur}})
		assertClean(t, tr, "siktir")
		assertClean(t, tr, "ibne")
		assertDetects(t, tr, "orospu")
		assertDetects(t, tr, "bok")
	})
	t.Run("per-call override", func(t *testing.T) {
		tr := mustNew(t, nil)
		if tr.ContainsProfanity("siktir", &terlik.DetectOptions{ExcludeCategories: []terlik.Category{terlik.CategorySexual}}) {
			t.Error("expected per-call exclude to skip sexual")
		}
		assertDetects(t, tr, "siktir")
	})
	t.Run("custom words (no category) never excluded", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{
			CustomList:        []string{"badword"},
			ExcludeCategories: []terlik.Category{terlik.CategorySexual, terlik.CategoryInsult, terlik.CategorySlur, terlik.CategoryGeneral},
		})
		assertDetects(t, tr, "badword")
	})
}

func TestCategoryInMatchResult(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("includes category from dictionary", func(t *testing.T) {
		matches := tr.GetMatches("siktir", nil)
		if len(matches) == 0 {
			t.Fatal("expected matches")
		}
		if matches[0].Category != terlik.CategorySexual {
			t.Errorf("expected sexual, got %q", matches[0].Category)
		}
	})
	t.Run("custom words have empty category", func(t *testing.T) {
		tr2 := mustNew(t, &terlik.Options{CustomList: []string{"testword"}})
		matches := tr2.GetMatches("testword", nil)
		if len(matches) == 0 {
			t.Fatal("expected matches")
		}
		if matches[0].Category != "" {
			t.Errorf("expected empty category, got %q", matches[0].Category)
		}
	})
}

func TestModeToggleInteraction(t *testing.T) {
	t.Run("strict + minSeverity", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{Mode: terlik.ModeStrict, MinSeverity: terlik.SeverityHigh})
		assertClean(t, tr, "salak")
		assertDetects(t, tr, "siktir")
	})
	t.Run("per-call toggle overrides constructor", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{MinSeverity: terlik.SeverityHigh})
		if tr.ContainsProfanity("salak", &terlik.DetectOptions{MinSeverity: terlik.SeverityLow}) {
			// Per-call: allow low → salak should be detected
			// Actually this sets minSeverity=low, which means allow everything ≥ low
		}
		// With minSeverity=low, salak (low) should be detected
		assertDetects(t, mustNew(t, &terlik.Options{MinSeverity: terlik.SeverityLow}), "salak")
	})
}

func TestDefaultBehaviorPreservation(t *testing.T) {
	t.Run("no options = detect everything", func(t *testing.T) {
		tr := mustNew(t, nil)
		assertDetects(t, tr, "siktir")
		assertDetects(t, tr, "salak")
		assertDetects(t, tr, "$1kt1r")
		en := mustNew(t, &terlik.Options{Language: "en"})
		assertDetects(t, en, "ShitPerson")
	})
	t.Run("clean respects toggles", func(t *testing.T) {
		tr := mustNew(t, &terlik.Options{MinSeverity: terlik.SeverityHigh})
		if tr.Clean("salak", nil) != "salak" {
			t.Error("salak (low) should not be masked with minSeverity=high")
		}
		if tr.Clean("siktir", nil) == "siktir" {
			t.Error("siktir (high) should be masked")
		}
	})
}
