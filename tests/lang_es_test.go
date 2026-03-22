package terlik_test

import (
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func TestSpanishRootDetection(t *testing.T) {
	es := mustNew(t, &terlik.Options{Language: "es"})

	roots := []struct{ word, text string }{
		{"mierda", "eso es una mierda"}, {"puta", "hijo de puta"}, {"cabron", "eres un cabron"},
		{"joder", "joder tio"}, {"coño", "coño ya"}, {"verga", "vete a la verga"},
		{"chingar", "no me chingar"}, {"pendejo", "eres pendejo"}, {"marica", "no seas marica"},
		{"carajo", "vete al carajo"}, {"idiota", "eres idiota"}, {"culo", "mueve el culo"},
		{"zorra", "esa zorra"}, {"estupido", "eres estupido"}, {"imbecil", "pedazo de imbecil"},
		{"gilipollas", "menudo gilipollas"}, {"huevon", "que huevon"}, {"pinche", "pinche idiota"},
		{"culero", "eres culero"}, {"cojones", "tiene cojones"}, {"polla", "menuda polla"},
		{"follar", "vamos a follar"}, {"capullo", "eres un capullo"}, {"guarro", "que guarro"},
		{"boludo", "sos un boludo"}, {"pelotudo", "que pelotudo"},
		{"hostia", "hostia tio"}, {"soplapollas", "menudo soplapollas"},
	}
	for _, r := range roots {
		t.Run("detects "+r.word, func(t *testing.T) { assertDetects(t, es, r.text) })
	}
}

func TestSpanishVariantDetection(t *testing.T) {
	es := mustNew(t, &terlik.Options{Language: "es"})

	variants := []string{
		"puto", "putas", "hijoputa", "putear", "putazo", "puteada",
		"jodido", "jodida", "jodiendo",
		"chingado", "chingada", "chingon", "chingona", "chingadera", "chingue",
		"pendejos", "pendeja", "pendejada", "maricon", "maricones",
		"cabrones", "cabrona", "cabronazo", "mierdoso",
		"estupida", "estupidez", "coñazo",
		"culeros", "culera", "cojonudo",
		"pollas", "pollon", "follando", "follado", "follada",
		"capullos", "guarros", "guarra", "guarrada",
		"boludos", "boluda", "boludez",
		"pelotudos", "pelotuda", "pelotudez", "mamonazo",
	}
	for _, v := range variants {
		t.Run("detects "+v, func(t *testing.T) { assertDetects(t, es, v) })
	}
}

func TestSpanishEvasion(t *testing.T) {
	es := mustNew(t, &terlik.Options{Language: "es"})
	t.Run("separator m.i.e.r.d.a", func(t *testing.T) { assertDetects(t, es, "m.i.e.r.d.a") })
	t.Run("leet m1erda", func(t *testing.T) { assertDetects(t, es, "m1erda") })
	t.Run("separator p.u.t.a", func(t *testing.T) { assertDetects(t, es, "p.u.t.a") })
}

func TestSpanishWhitelist(t *testing.T) {
	es := mustNew(t, &terlik.Options{Language: "es"})
	safeWords := []string{
		"computadora", "disputar", "reputacion", "calcular",
		"particular", "vehicular", "pollo", "pollito", "polleria", "polluelo",
		"folleto", "follaje",
	}
	for _, word := range safeWords {
		t.Run("safe: "+word, func(t *testing.T) { assertClean(t, es, word) })
	}
}

func TestSpanishIsolation(t *testing.T) {
	es := mustNew(t, &terlik.Options{Language: "es"})
	assertClean(t, es, "siktir git")
	assertClean(t, es, "what the fuck")
}
