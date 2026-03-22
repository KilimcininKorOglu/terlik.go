package terlik_test

import (
	"strings"
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
	"time"
)

// Go's RE2 engine guarantees O(n) time — no catastrophic backtracking.
// These tests verify that adversarial inputs complete within reasonable time.
const maxDetectMs = 30000 // generous for CI

func TestReDoSAdversarialTiming(t *testing.T) {
	tr := mustNew(t, nil)
	en := mustNew(t, &terlik.Options{Language: "en"})

	// Warm up
	tr.ContainsProfanity("warmup siktir", nil)
	en.ContainsProfanity("warmup fuck", nil)

	t.Run("repeated separator characters", func(t *testing.T) {
		input := "a" + strings.Repeat(".", 100) + "b" + strings.Repeat(".", 100) + "c"
		start := time.Now()
		tr.ContainsProfanity(input, nil)
		if time.Since(start).Milliseconds() > maxDetectMs {
			t.Error("took too long")
		}
	})
	t.Run("long overlap @ signs", func(t *testing.T) {
		input := strings.Repeat("@", 50)
		start := time.Now()
		tr.ContainsProfanity(input, nil)
		if time.Since(start).Milliseconds() > maxDetectMs {
			t.Error("took too long")
		}
	})
	t.Run("long overlap $ signs", func(t *testing.T) {
		input := strings.Repeat("$", 50)
		start := time.Now()
		en.ContainsProfanity(input, nil)
		if time.Since(start).Milliseconds() > maxDetectMs {
			t.Error("took too long")
		}
	})
	t.Run("maximum length input 10K chars", func(t *testing.T) {
		input := strings.Repeat("test", 2500)
		start := time.Now()
		tr.ContainsProfanity(input, nil)
		if time.Since(start).Milliseconds() > maxDetectMs {
			t.Error("took too long")
		}
	})
	t.Run("leet + separator adversarial combo", func(t *testing.T) {
		input := "$" + strings.Repeat("...", 20) + "1" + strings.Repeat("...", 20) + "k"
		start := time.Now()
		tr.ContainsProfanity(input, nil)
		if time.Since(start).Milliseconds() > maxDetectMs {
			t.Error("took too long")
		}
	})
}

func TestReDoSAttackSurface(t *testing.T) {
	tr := mustNew(t, nil)
	en := mustNew(t, &terlik.Options{Language: "en"})

	t.Run("separator abuse", func(t *testing.T) {
		assertDetects(t, tr, "s.i.k")
		assertDetects(t, tr, "s-i-k")
		assertDetects(t, tr, "s_i_k")
		assertDetects(t, tr, "s*i*k")
		assertDetects(t, en, "f.u.c.k")
		assertDetects(t, en, "f-u-c-k")
	})
	t.Run("repeat abuse", func(t *testing.T) {
		assertDetects(t, tr, "siiiiik")
		assertDetects(t, tr, "siiiiiktir")
		assertDetects(t, en, "fuuuuck")
	})
	t.Run("mixed evasion", func(t *testing.T) {
		assertDetects(t, tr, "$1kt1r")
		assertDetects(t, en, "ph.u.c.k")
	})
	t.Run("unicode normalization", func(t *testing.T) {
		assertDetects(t, tr, "ｓｉｋｔｉｒ")
		assertDetects(t, en, "ｆｕｃｋ")
	})
	t.Run("zero-width char evasion", func(t *testing.T) {
		assertDetects(t, tr, "s\u200Bi\u200Bk\u200Bt\u200Bi\u200Br")
		assertDetects(t, en, "f\u200Bu\u200Bc\u200Bk")
	})
	t.Run("Cyrillic confusable", func(t *testing.T) {
		assertDetects(t, en, "fu\u0441k")
	})
	t.Run("zalgo text", func(t *testing.T) {
		assertDetects(t, tr, "s\u0300i\u0301k\u0302t\u0303i\u0304r")
	})
}

func TestReDoSFalsePositivePrevention(t *testing.T) {
	tr := mustNew(t, nil)
	en := mustNew(t, &terlik.Options{Language: "en"})

	trClean := []string{
		"amsterdam güzel bir şehir", "malzeme listesi", "ameliyat günü",
		"sıkıntı yok", "bokser köpek cinsi", "memur kadrosu",
		"sikke koleksiyonu", "amir geldi", "dolunay vardi",
	}
	for _, w := range trClean {
		t.Run("TR clean: "+w, func(t *testing.T) { assertClean(t, tr, w) })
	}

	enClean := []string{
		"assassin movie", "class reunion", "cocktail party",
		"grasshopper jump", "passage through", "assessment report",
		"piston engine", "massage parlor", "shiitake mushroom",
	}
	for _, w := range enClean {
		t.Run("EN clean: "+w, func(t *testing.T) { assertClean(t, en, w) })
	}
}
