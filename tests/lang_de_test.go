package terlik_test

import (
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func TestGermanRootDetection(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})

	roots := []struct{ word, text string }{
		{"scheiße", "das ist scheiße"}, {"fick", "fick dich"}, {"arsch", "du arsch"},
		{"hurensohn", "hurensohn"}, {"hure", "du hure"}, {"fotze", "blöde fotze"},
		{"wichser", "du wichser"}, {"schwanz", "leck meinen schwanz"}, {"schlampe", "du schlampe"},
		{"mistkerl", "so ein mistkerl"}, {"idiot", "du idiot"}, {"dumm", "bist du dumm"},
		{"depp", "du depp"}, {"vollidiot", "so ein vollidiot"}, {"missgeburt", "du missgeburt"},
		{"drecksau", "du drecksau"}, {"dreck", "so ein dreck"}, {"trottel", "du trottel"},
		{"schwuchtel", "du schwuchtel"}, {"spast", "du spast"}, {"miststück", "du miststück"},
		{"bastard", "du bastard"}, {"penner", "du penner"}, {"blödmann", "du blödmann"},
		{"vollpfosten", "so ein vollpfosten"}, {"hackfresse", "du hackfresse"},
		{"pissnelke", "du pissnelke"}, {"spacken", "du spacken"},
	}
	for _, r := range roots {
		t.Run("detects "+r.word, func(t *testing.T) { assertDetects(t, de, r.text) })
	}
}

func TestGermanVariantDetection(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})

	variants := []string{
		"scheisse", "scheiss", "beschissen", "scheissegal",
		"ficken", "ficker", "gefickt", "verfickt", "fickfehler",
		"arschloch", "arschgeige", "arschgesicht", "arschbacke", "arschlocher",
		"fotzen", "wichsen", "gewichst", "wixer",
		"schlampig", "schlamperei", "dummkopf", "dummheit",
		"dreckig", "drecksack", "vollidioten", "missgeburten",
		"schwuchteln", "spasten", "spasti", "miststueck",
		"bastarde", "blodmann",
	}
	for _, v := range variants {
		t.Run("detects "+v, func(t *testing.T) { assertDetects(t, de, v) })
	}
}

func TestGermanSzHandling(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})
	t.Run("detects scheiße with ß", func(t *testing.T) { assertDetects(t, de, "scheiße") })
	t.Run("detects scheisse without ß", func(t *testing.T) { assertDetects(t, de, "scheisse") })
	t.Run("detects SCHEISSE uppercase", func(t *testing.T) { assertDetects(t, de, "SCHEISSE") })
}

func TestGermanEvasion(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})
	t.Run("separator f.i.c.k", func(t *testing.T) { assertDetects(t, de, "f.i.c.k") })
	t.Run("separator s.c.h.e.i.s.s.e", func(t *testing.T) { assertDetects(t, de, "s.c.h.e.i.s.s.e") })
}

func TestGermanWhitelist(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})
	safeWords := []string{"schwanger", "schwangerschaft", "geschichte"}
	for _, word := range safeWords {
		t.Run("safe: "+word, func(t *testing.T) { assertClean(t, de, word) })
	}
}

func TestGermanCleanText(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})
	assertClean(t, de, "hallo welt wie geht es dir")
}

func TestGermanIsolation(t *testing.T) {
	de := mustNew(t, &terlik.Options{Language: "de"})
	assertClean(t, de, "siktir git")
	assertClean(t, de, "what the fuck")
	assertClean(t, de, "hijo de puta")
}
