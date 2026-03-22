package terlik_test

import "testing"

// Comprehensive profanity detection tests — all roots
// Each root: plain text, variant, in sentence, uppercase, suffix, whitelist

func TestProfanitySik(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain root", "sik"}, {"variant siktir", "siktir git"}, {"variant sikerim", "sikerim seni"},
		{"variant sikicem", "sikicem"}, {"variant siktim", "siktim"}, {"variant sikeyim", "sikeyim"},
		{"variant sikis", "sikis"}, {"variant sikik", "sikik herif"}, {"variant sikim", "sikim"},
		{"Turkish İ", "SİKTİR"}, {"leet $1kt1r", "$1kt1r"}, {"separator s.i.k", "s.i.k"},
		{"repeat siiiiiktir", "siiiiiktir"}, {"in sentence", "hadi siktir git burdan"},
		{"suffix siktiler", "siktiler"}, {"suffix siktirler", "siktirler"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "sik") })
	}
	t.Run("whitelist sikke", func(t *testing.T) { assertClean(t, tr, "sikke") })
	t.Run("whitelist siklet", func(t *testing.T) { assertClean(t, tr, "siklet") })
}

func TestProfanityAmk(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain amk", "amk"}, {"variant aminakoyim", "aminakoyim"},
		{"variant aminakoydugum", "aminakoydugum"}, {"variant amq", "amq"},
		{"in sentence", "bu ne amk"}, {"uppercase", "AMK"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "amk") })
	}
	// "amina" is a variant of both "am" and "amk" — just verify detection
	t.Run("variant amina", func(t *testing.T) { assertDetects(t, tr, "amina koyarim") })
}

func TestProfanityOrospu(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain", "orospu"}, {"variant orospucocugu", "orospucocugu"},
		{"variant orspu", "orspu"}, {"variant oruspu", "oruspu"}, {"variant orosbu", "orosbu"},
		{"in sentence", "kes lan orospu cocugu"}, {"suffix orospuluk", "orospuluk yapma"},
		{"suffix orospular", "orospular"}, {"separator", "o.r.o.s.p.u"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "orospu") })
	}
}

func TestProfanityPic(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain piç", "piç kurusu"}, {"normalized pic", "pic"},
		{"variant piclik", "piclik yapma"}, {"in sentence", "bu adam pic"}, {"suffix picler", "picler"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "piç") })
	}
	t.Run("whitelist piknik", func(t *testing.T) { assertClean(t, tr, "piknik") })
	t.Run("whitelist pikachu", func(t *testing.T) { assertClean(t, tr, "pikachu") })
}

func TestProfanityYarrak(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain yarrak", "yarrak"}, {"variant yarak", "yarak"}, {"variant yarakli", "yarakli"},
		{"variant dalyarak", "dalyarak"}, {"variant dalyarrak", "dalyarrak"},
		{"in sentence", "bu ne yarrak gibi is"}, {"suffix yarraklar", "yarraklar"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "yarrak") })
	}
	t.Run("whitelist yarasa", func(t *testing.T) { assertClean(t, tr, "yarasa") })
}

func TestProfanityGot(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain göt", "göt"}, {"normalized got", "got"}, {"variant gotunu", "gotunu"},
		{"variant gotlek", "gotlek"}, {"variant gotveren", "gotveren"},
		{"variant gotverenler", "gotverenler"}, {"in sentence", "senin got"},
		{"leet 6öt", "6öt"}, {"suffix gotune", "gotune"}, {"suffix gotler", "gotler"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "göt") })
	}
	t.Run("whitelist gotik", func(t *testing.T) { assertClean(t, tr, "gotik") })
	t.Run("whitelist gotham", func(t *testing.T) { assertClean(t, tr, "gotham") })
}

func TestProfanityAm(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"variant amcik", "amcik"}, {"variant amcuk", "amcuk"},
		{"in sentence", "amcik herif"}, {"variant ami", "ami"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "am") })
	}
	for _, w := range []string{"ama", "amen", "amir", "ambalaj", "ambulans", "ameliyat", "amerika", "amino", "amonyak", "ampul"} {
		t.Run("whitelist "+w, func(t *testing.T) { assertClean(t, tr, w) })
	}
}

func TestProfanityTasak(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain taşak", "taşak"}, {"normalized tasak", "tasak"}, {"variant tassak", "tassak"},
		{"variant tassakli", "tassakli"}, {"in sentence", "tasak gecme"}, {"suffix tasaklar", "tasaklar"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "taşak") })
	}
}

func TestProfanityMeme(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain meme", "meme"}, {"in sentence", "meme gosterdi"}, {"uppercase", "MEME"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "meme") })
	}
	for _, w := range []string{"memento", "memleket", "memur", "memorial"} {
		t.Run("whitelist "+w, func(t *testing.T) { assertClean(t, tr, w) })
	}
}

func TestProfanityIbne(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain ibne", "ibne"}, {"variant ibneler", "ibneler"}, {"in sentence", "lan ibne"},
		{"leet i8ne", "i8ne"}, {"suffix ibnelik", "ibnelik"}, {"suffix ibneler", "ibneler geldi"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "ibne") })
	}
}

func TestProfanityGavat(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain gavat", "gavat"}, {"variant gavatlik", "gavatlik"},
		{"in sentence", "bu adam gavat"}, {"suffix gavatlar", "gavatlar"}, {"uppercase", "GAVAT"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "gavat") })
	}
}

func TestProfanityPezevenk(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain pezevenk", "pezevenk"}, {"variant pezo", "pezo herif"},
		{"in sentence", "bu pezevenk kim"}, {"suffix pezevenkler", "pezevenkler"},
		{"suffix pezevenklik", "pezevenklik"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "pezevenk") })
	}
}

func TestProfanityBok(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain bok", "bok"}, {"variant boktan", "boktan"},
		{"in sentence", "ne boktan bir gun"}, {"leet 8ok", "8ok"},
		{"suffix boklar", "boklar"}, {"suffix boklu", "boklu"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "bok") })
	}
	t.Run("whitelist bokser", func(t *testing.T) { assertClean(t, tr, "bokser") })
	t.Run("whitelist boksör", func(t *testing.T) { assertClean(t, tr, "boksör") })
}

func TestProfanitySalak(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain salak", "salak"}, {"variant salaklik", "salaklik"},
		{"in sentence", "salak misin sen"}, {"uppercase", "SALAK"},
		{"suffix salaksin", "salaksin"}, {"suffix salaklar", "salaklar"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "salak") })
	}
}

func TestProfanityAptal(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain aptal", "aptal"}, {"variant aptallik", "aptallik"}, {"variant aptalca", "aptalca"},
		{"in sentence", "bu adam aptal herif"}, {"leet @pt@l", "@pt@l"},
		{"suffix aptallar", "aptallar"}, {"suffix aptallarin", "aptallarin isi"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "aptal") })
	}
}

func TestProfanityGerizekali(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain gerizekalı", "gerizekalı"}, {"normalized gerizekali", "gerizekali"},
		{"in sentence", "bu gerizekali kim"}, {"suffix gerizekaliler", "gerizekaliler"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "gerizekalı") })
	}
}

func TestProfanityMal(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain mal", "mal herif"}, {"in sentence", "bu adam mal"}, {"uppercase", "MAL"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "mal") })
	}
	for _, w := range []string{"malzeme", "maliyet", "malik", "malikane", "maliye", "malta", "malt", "mallorca"} {
		t.Run("whitelist "+w, func(t *testing.T) { assertClean(t, tr, w) })
	}
}

func TestProfanityDol(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain döl", "döl"}, {"normalized dol", "dol"},
		{"variant dolunu", "dolunu"}, {"in sentence", "dol israfı"}, {"variant dolcu", "dolcu"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "döl") })
	}
	for _, w := range []string{"dolunay", "dolum", "doluluk", "dolmen"} {
		t.Run("whitelist "+w, func(t *testing.T) { assertClean(t, tr, w) })
	}
}

func TestProfanityKahpe(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain kahpe", "kahpe"}, {"variant kahpelik", "kahpelik"},
		{"in sentence", "kahpe kari"}, {"suffix kahpeler", "kahpeler"},
		{"suffix kahpelikler", "kahpelikler"}, {"uppercase", "KAHPE"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "kahpe") })
	}
}

func TestProfanitySurtuk(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain sürtük", "sürtük"}, {"normalized surtuk", "surtuk"},
		{"in sentence", "bu kadin surtuk"}, {"suffix surtukler", "surtukler"},
		{"suffix surtukluk", "surtukluk"}, {"uppercase", "SÜRTÜK"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "sürtük") })
	}
}

func TestProfanityKaltak(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain kaltak", "kaltak"}, {"in sentence", "bu kaltak kim"},
		{"suffix kaltaklar", "kaltaklar"}, {"suffix kaltaklik", "kaltaklik"}, {"uppercase", "KALTAK"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "kaltak") })
	}
}

func TestProfanityFahise(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain fahişe", "fahişe"}, {"normalized fahise", "fahise"},
		{"in sentence", "bu fahise kadin"}, {"suffix fahiseler", "fahiseler"},
		{"suffix fahiselik", "fahiselik"}, {"uppercase", "FAHISE"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "fahişe") })
	}
}

func TestProfanityPust(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain puşt", "puşt"}, {"normalized pust", "pust"}, {"variant pustt", "pustt"},
		{"in sentence", "lan pust"}, {"leet pu$t", "pu$t"},
		{"suffix pustlar", "pustlar"}, {"suffix pustluk", "pustluk"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "puşt") })
	}
}

func TestProfanitySerefsiz(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain şerefsiz", "şerefsiz"}, {"normalized serefsiz", "serefsiz"},
		{"variant serefsizler", "serefsizler"}, {"in sentence", "bu serefsiz adam"},
		{"suffix serefsizlik", "serefsizlik"}, {"uppercase", "SEREFSIZ"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "şerefsiz") })
	}
}

func TestProfanityYavsak(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain yavşak", "yavşak"}, {"normalized yavsak", "yavsak"},
		{"in sentence", "bu yavsak kim"}, {"suffix yavsaklik", "yavsaklik"},
		{"suffix yavsaklar", "yavsaklar"}, {"uppercase", "YAVSAK"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "yavşak") })
	}
}

func TestProfanityNamussuz(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain namussuz", "namussuz"}, {"in sentence", "bu namussuz adam"},
		{"suffix namussuzlar", "namussuzlar"}, {"suffix namussuzluk", "namussuzluk"}, {"uppercase", "NAMUSSUZ"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "namussuz") })
	}
	t.Run("whitelist namus", func(t *testing.T) { assertClean(t, tr, "namus") })
	t.Run("whitelist namuslu", func(t *testing.T) { assertClean(t, tr, "namuslu") })
}

func TestProfanityAhlaksiz(t *testing.T) {
	tr := mustNew(t, nil)
	for _, tt := range []struct{ d, i string }{
		{"plain ahlaksız", "ahlaksız"}, {"normalized ahlaksiz", "ahlaksiz"},
		{"in sentence", "bu ahlaksiz adam"}, {"suffix ahlaksizlar", "ahlaksizlar"},
		{"suffix ahlaksizlik", "ahlaksizlik"}, {"uppercase", "AHLAKSIZ"},
	} {
		t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, "ahlaksız") })
	}
	t.Run("whitelist ahlak", func(t *testing.T) { assertClean(t, tr, "ahlak") })
	t.Run("whitelist ahlaki", func(t *testing.T) { assertClean(t, tr, "ahlaki") })
}

func TestProfanityRemainingRoots(t *testing.T) {
	tr := mustNew(t, nil)

	// haysiyetsiz, dangalak, ezik, dingil, avanak, manyak, hödük, kepaze, rezil, kalleş, kevaşe, oğlancı
	simpleRoots := []struct{ root string; cases []struct{ d, i string } }{
		{"haysiyetsiz", []struct{ d, i string }{{"plain", "haysiyetsiz"}, {"in sentence", "bu adam haysiyetsiz"}, {"uppercase", "HAYSIYETSIZ"}}},
		{"dangalak", []struct{ d, i string }{{"plain", "dangalak"}, {"in sentence", "bu dangalak ne yapiyor"}, {"suffix", "dangalaklar"}, {"uppercase", "DANGALAK"}}},
		{"ezik", []struct{ d, i string }{{"plain", "ezik"}, {"in sentence", "ezik herif"}, {"suffix ezikler", "ezikler"}, {"suffix eziklik", "eziklik"}, {"uppercase", "EZIK"}}},
		{"dingil", []struct{ d, i string }{{"plain", "dingil"}, {"in sentence", "bu dingil ne yapiyor"}, {"suffix dingiller", "dingiller"}, {"suffix dingillik", "dingillik"}, {"uppercase", "DINGIL"}}},
		{"avanak", []struct{ d, i string }{{"plain", "avanak"}, {"in sentence", "bu avanak herif"}, {"suffix avanaklar", "avanaklar"}, {"suffix avanaklik", "avanaklik"}, {"uppercase", "AVANAK"}}},
		{"manyak", []struct{ d, i string }{{"plain", "manyak"}, {"in sentence", "bu adam manyak"}, {"suffix manyaklar", "manyaklar"}, {"suffix manyaklik", "manyaklik"}, {"uppercase", "MANYAK"}}},
		{"hödük", []struct{ d, i string }{{"plain hödük", "hödük"}, {"normalized hoduk", "hoduk"}, {"in sentence", "bu hoduk herif"}, {"suffix hodukler", "hodukler"}, {"suffix hodukluk", "hodukluk"}, {"uppercase", "HODUK"}}},
		{"kepaze", []struct{ d, i string }{{"plain", "kepaze"}, {"in sentence", "bu adam kepaze"}, {"suffix kepazeler", "kepazeler"}, {"suffix kepazelik", "kepazelik"}, {"uppercase", "KEPAZE"}}},
		{"rezil", []struct{ d, i string }{{"plain", "rezil"}, {"in sentence", "rezil oldu"}, {"suffix reziller", "reziller"}, {"suffix rezillik", "rezillik"}, {"uppercase", "REZIL"}}},
		{"kalleş", []struct{ d, i string }{{"plain kalleş", "kalleş"}, {"normalized kalles", "kalles"}, {"in sentence", "bu kalles herif"}, {"suffix kallesler", "kallesler"}, {"suffix kalleslik", "kalleslik"}, {"uppercase", "KALLES"}}},
		{"kevaşe", []struct{ d, i string }{{"plain kevaşe", "kevaşe"}, {"normalized kevase", "kevase"}, {"in sentence", "bu kevase kim"}, {"suffix kevaseler", "kevaseler"}, {"uppercase", "KEVASE"}}},
		{"oğlancı", []struct{ d, i string }{{"plain oğlancı", "oğlancı"}, {"normalized oglanci", "oglanci"}, {"in sentence", "bu adam oglanci"}, {"suffix oglancilar", "oglancilar"}, {"suffix oglancilik", "oglancilik"}, {"uppercase", "OGLANCI"}}},
	}

	for _, r := range simpleRoots {
		t.Run("root: "+r.root, func(t *testing.T) {
			for _, tt := range r.cases {
				t.Run(tt.d, func(t *testing.T) { assertDetectsRoot(t, tr, tt.i, r.root) })
			}
		})
	}
}
