package terlik_test

import (
	"testing"
)

func TestSuffixEngine(t *testing.T) {
	tr := mustNew(t, nil)

	t.Run("suffixable roots catch suffixed forms", func(t *testing.T) {
		cases := []struct{ input, desc string }{
			{"siktiler hepsini", "siktiler"},
			{"sikerim seni", "sikerim"},
			{"orospuluk yapma", "orospuluk"},
			{"gotune sokayim", "gotune"},
			{"boktan bir gun", "boktan"},
			{"ibnelik yapma", "ibnelik"},
			{"gavatlar geldi", "gavatlar"},
			{"salaksin sen", "salaksin"},
			{"aptallarin isi", "aptallarin"},
			{"kahpeler burada", "kahpeler"},
			{"pezevenkler toplandi", "pezevenkler"},
			{"yavsaklik etme", "yavsaklik"},
			{"serefsizler", "serefsizler"},
			{"pustlar geldi", "pustlar"},
		}
		for _, tt := range cases {
			t.Run(tt.desc, func(t *testing.T) {
				assertDetects(t, tr, tt.input)
			})
		}
	})

	t.Run("suffix chaining up to 2", func(t *testing.T) {
		assertDetects(t, tr, "siktirler hep")
		assertDetects(t, tr, "orospuluklar")
	})

	t.Run("evasion + suffix", func(t *testing.T) {
		assertDetects(t, tr, "s.i.k.t.i.r.l.e.r")
		assertDetects(t, tr, "$1kt1rler")
	})

	t.Run("non-suffixable entries reject suffix forms", func(t *testing.T) {
		assertClean(t, tr, "ama neden")
		assertDetects(t, tr, "ami bozuk")
	})

	t.Run("false positive prevention", func(t *testing.T) {
		cleanWords := []string{
			"ama ben istemiyorum", "amen dedi", "osmanlı sikke",
			"amsterdam", "bokser kopek cinsi", "dolmen yapisi",
			"dolunay vardi", "sıkma limon", "sıkıntı var",
			"sıkıştı araba", "sıkı çalış", "amir geldi",
		}
		for _, w := range cleanWords {
			t.Run(w, func(t *testing.T) {
				assertClean(t, tr, w)
			})
		}
		assertDetects(t, tr, "ami problemi var")
	})
}
