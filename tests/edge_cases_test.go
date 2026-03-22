package terlik_test

import (
	"strings"
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func TestFalsePositivePrevention(t *testing.T) {
	tr := mustNew(t, nil)

	cleanWords := []string{
		"osmanlı sikke koleksiyonu", "amsterdam güzel şehir",
		"ambulans geldi", "ameliyat olacak", "malzeme listesi",
		"devlet memuru", "bokser köpek cinsi", "ama yapamam",
		"amen dedi papaz", "amir bey geldi", "dolmen antik yapi",
		"amazon siparis verdim", "ambargo uygulandi", "amblem tasarimi",
		"amfibi arac", "dolap kapagi", "dolar kuru yukseldi",
		"dolma yaprak sardim", "dolmus bekliyorum", "malum kisi",
		"namus meselesi", "namuslu adam", "ahlak dersi", "ahlaki degerler",
	}
	for _, w := range cleanWords {
		t.Run(w, func(t *testing.T) {
			assertClean(t, tr, w)
		})
	}

	t.Run("flags ami (am variant)", func(t *testing.T) {
		assertDetects(t, tr, "ami sorunu")
	})
}

func TestEmojiHandling(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("detects profanity with surrounding emojis", func(t *testing.T) {
		assertDetects(t, tr, "😡 siktir 😡")
	})
	t.Run("does not false-positive on emoji-only text", func(t *testing.T) {
		assertClean(t, tr, "😀😁😂🤣")
	})
}

func TestLongInput(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("handles input up to maxLength", func(t *testing.T) {
		longClean := strings.Repeat("merhaba ", 2000)
		assertClean(t, tr, longClean)
	})
	t.Run("truncates input beyond maxLength", func(t *testing.T) {
		tr2 := mustNew(t, &terlik.Options{MaxLength: 20})
		text := strings.Repeat("a", 25) + " siktir"
		assertClean(t, tr2, text)
	})
}

func TestEmptyAndSpecialInputs(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("handles empty string", func(t *testing.T) {
		assertClean(t, tr, "")
		if tr.Clean("", nil) != "" {
			t.Error("clean('') should return ''")
		}
		if len(tr.GetMatches("", nil)) != 0 {
			t.Error("getMatches('') should return empty")
		}
	})
	t.Run("handles whitespace only", func(t *testing.T) {
		assertClean(t, tr, "   ")
	})
	t.Run("handles numbers only", func(t *testing.T) {
		assertClean(t, tr, "123456")
	})
	t.Run("handles special characters only", func(t *testing.T) {
		assertClean(t, tr, "!@#$%^&*()")
	})
}

func TestTurkishCharacterVariations(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("detects with Turkish İ/ı", func(t *testing.T) {
		assertDetects(t, tr, "SİKTİR")
	})
	t.Run("detects with mixed case Turkish", func(t *testing.T) {
		assertDetects(t, tr, "Sİktİr")
	})
}

func TestLeetSpeakEvasion(t *testing.T) {
	tr := mustNew(t, nil)

	cases := []struct{ input, desc string }{
		{"$1kt1r lan", "$1kt1r"},
		{"@pt@l herif", "@pt@l"},
		{"8ok herif", "8ok (bok)"},
		{"senin 6öt", "6öt (göt)"},
		{"s2kt2r", "s2kt2r (pattern charClass)"},
	}
	for _, tt := range cases {
		t.Run(tt.desc, func(t *testing.T) {
			assertDetects(t, tr, tt.input)
		})
	}
}

func TestCharacterRepetitionEvasion(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("detects siiiiiktir", func(t *testing.T) {
		assertDetects(t, tr, "siiiiiktir")
	})
	t.Run("detects orrrospu", func(t *testing.T) {
		assertDetects(t, tr, "orrrospu")
	})
}

func TestSeparatorEvasion(t *testing.T) {
	tr := mustNew(t, nil)

	for _, sep := range []string{".", "-", "_"} {
		input := "s" + sep + "i" + sep + "k" + sep + "t" + sep + "i" + sep + "r"
		t.Run("detects s"+sep+"i"+sep+"k"+sep+"t"+sep+"i"+sep+"r", func(t *testing.T) {
			assertDetects(t, tr, input)
		})
	}
}

func TestNewVariantDetection(t *testing.T) {
	tr := mustNew(t, nil)

	variants := []struct{ input, desc string }{
		{"aminakoyayim", "aminakoyayim"},
		{"aminakoydum", "aminakoydum"},
		{"aminakoydugumun", "aminakoydugumun"},
		{"bu ne aq", "aq"},
		{"orospucocuklari", "orospucocuklari"},
		{"gotos herif", "gotos"},
		{"yarrani ye", "yarrani"},
		{"yarragimi", "yarragimi"},
		{"yarragini", "yarragini"},
		{"sktr lan", "sktr"},
	}
	for _, tt := range variants {
		t.Run(tt.desc, func(t *testing.T) {
			assertDetects(t, tr, tt.input)
		})
	}
}

func TestTurkishNumberEvasion(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("detects s2k (sikik)", func(t *testing.T) {
		assertDetects(t, tr, "s2k herif")
	})
	t.Run("detects s2mle (sikimle)", func(t *testing.T) {
		assertDetects(t, tr, "s2mle ugras")
	})
	t.Run("does not flag standalone numbers", func(t *testing.T) {
		assertClean(t, tr, "2023 yilinda")
		assertClean(t, tr, "skor 2-1 oldu")
	})
	t.Run("does not flag 100 kisi", func(t *testing.T) {
		assertClean(t, tr, "100 kisi geldi")
	})
}
