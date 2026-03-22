package terlik_test

import (
	"terlik"
	"testing"
)

func TestEnglishRootDetection(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	roots := []struct{ word, text string }{
		{"fuck", "what the fuck"}, {"shit", "this is shit"}, {"bitch", "son of a bitch"},
		{"damn", "damn it"}, {"asshole", "what an asshole"}, {"dick", "don't be a dick"},
		{"cock", "what a cock"}, {"cunt", "you cunt"}, {"whore", "dirty whore"},
		{"slut", "she is a slut"}, {"piss", "piss off"}, {"wank", "go wank"},
		{"twat", "you twat"}, {"bollocks", "that is bollocks"}, {"crap", "what crap"},
		{"retard", "you retard"}, {"faggot", "stupid faggot"}, {"douche", "total douche"},
		{"spic", "dirty spic"}, {"kike", "filthy kike"}, {"chink", "stupid chink"},
		{"gook", "dirty gook"}, {"tranny", "ugly tranny"}, {"dyke", "stupid dyke"},
		{"coon", "dirty coon"}, {"wetback", "filthy wetback"}, {"bellend", "you bellend"},
		{"skank", "total skank"}, {"scumbag", "what a scumbag"}, {"turd", "you turd"},
		{"bugger", "bugger off"}, {"hell", "go to hell"}, {"prick", "you prick"},
		{"screw", "screw you"}, {"porn", "watching porn"}, {"blowjob", "gave a blowjob"},
		{"jizz", "jizz everywhere"}, {"dildo", "bought a dildo"}, {"orgasm", "had an orgasm"},
		{"orgy", "wild orgy"}, {"hooker", "street hooker"}, {"negro", "dirty negro"},
		{"masturbate", "caught masturbating"}, {"semen", "covered in semen"},
		{"pussy", "wet pussy"}, {"cum", "wants cum"}, {"penis", "show me your penis"},
		{"tit", "nice tits"}, {"vagina", "lick my vagina"}, {"anal", "anal sex"},
		{"rape", "he raped her"},
	}
	for _, r := range roots {
		t.Run("detects "+r.word, func(t *testing.T) { assertDetects(t, en, r.text) })
	}
}

func TestEnglishVariantDetection(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	variants := []string{
		"fucking", "fucker", "motherfucker", "stfu", "fuckboy", "fucktard", "fuckhead", "wtf", "mofo",
		"unfucking", "fuckery", "shitty", "bullshit", "dipshit", "shithole", "shitbag", "shitload",
		"shithouse", "shitlist", "shitfaced", "bitchy", "bitching", "bitchslap",
		"cocksucker", "cocksucking", "cockblock", "slutty", "whorish",
		"pissed", "pissing", "wanker", "wanking", "retarded",
		"nigga", "fag", "fags", "douchebag", "dickhead", "dickwad",
		"jackass", "dumbass", "smartass", "asscrack", "assclown", "goddamn",
		"spicks", "kikes", "chinks", "chinky", "gooks",
		"trannies", "dykes", "coons", "wetbacks", "bellends", "skanky", "scumbags", "turds",
		"buggered", "buggering", "buggery",
		"hells", "pricks", "pricked", "pricking", "screwed", "screwing", "screws",
		"pornographic", "pornography", "porno", "blowjobs",
		"jizzed", "jizzing", "dildos", "orgasms", "orgasmic", "orgies", "hookers", "negroes",
		"masturbating", "masturbation", "pussies", "cumming", "cumshot",
		"penises", "tits", "titty", "titties", "vaginas", "vaginal",
		"raped", "raping", "rapist", "rapists",
		"fuckyou", "fuckoff", "fuckwad", "fuckup", "fuckall",
		"shitlord", "shitstain", "shitbrain", "cockwomble", "twatwaffle", "assmunch",
		"cumguzzler", "cumdumpster", "dickweasel", "thundercunt",
	}
	for _, v := range variants {
		t.Run("detects "+v, func(t *testing.T) { assertDetects(t, en, v) })
	}
}

func TestEnglishEvasionDetection(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	t.Run("separator f.u.c.k", func(t *testing.T) { assertDetects(t, en, "f.u.c.k") })
	t.Run("leet fck", func(t *testing.T) { assertDetects(t, en, "fck this") })
	t.Run("repetition fuuuck", func(t *testing.T) { assertDetects(t, en, "fuuuck") })
	t.Run("leet $h1t", func(t *testing.T) { assertDetects(t, en, "$h1t") })
	t.Run("leet b1tch", func(t *testing.T) { assertDetects(t, en, "b1tch") })
	t.Run("ph→f phuck", func(t *testing.T) { assertDetects(t, en, "phuck you") })
	t.Run("ph→f phucking", func(t *testing.T) { assertDetects(t, en, "phucking idiot") })
	t.Run("#→h s#it", func(t *testing.T) { assertDetects(t, en, "s#it stain") })
	t.Run("8→b 8itch", func(t *testing.T) { assertDetects(t, en, "8itch slap") })
	t.Run("6→g ni66er", func(t *testing.T) { assertDetects(t, en, "ni66er") })
	t.Run("combined n!66er", func(t *testing.T) { assertDetects(t, en, "n!66er") })
	t.Run("CamelCase FuckYou", func(t *testing.T) { assertDetects(t, en, "FuckYou") })
	t.Run("CamelCase ShitHead", func(t *testing.T) { assertDetects(t, en, "ShitHead") })
	t.Run("ALLCAPS+lower SHITlord", func(t *testing.T) { assertDetects(t, en, "SHITlord") })
	t.Run("hashtag #fuckyou", func(t *testing.T) { assertDetects(t, en, "#fuckyou") })
	t.Run("accented fück", func(t *testing.T) { assertDetects(t, en, "f\u00FCck") })
	t.Run("Cyrillic fuсk", func(t *testing.T) { assertDetects(t, en, "fu\u0441k") })
	t.Run("fullwidth ｆｕｃｋ", func(t *testing.T) { assertDetects(t, en, "\uFF46\uFF55\uFF43\uFF4B") })
}

func TestEnglishWhitelist(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})

	safeWords := []string{
		"assassin", "assassinate", "assistant", "assessment",
		"class", "classic", "classify", "classroom", "grass", "grasshopper",
		"mass", "massive", "pass", "passage", "passenger",
		"passion", "passive", "passport", "assume", "asset",
		"assess", "dickens", "cocktail", "cockatoo", "cockatiel",
		"cockpit", "cockroach", "cockney", "peacock", "shuttlecock", "woodcock",
		"scrap", "piston", "bassist", "embassy", "hassle",
		"massage", "compass", "harass", "shiitake",
		"cocoon", "raccoon", "tycoon", "dike", "vandyke", "scunthorpe",
		"cocked", "hello", "shell", "seashell", "eggshell", "nutshell",
		"bombshell", "helium", "helicopter", "helmet",
		"prickle", "prickly", "screwdriver", "corkscrew", "puck", "pucks",
		"pussycat", "pussywillow", "pussyfoot", "penistone",
		"analysis", "analyst", "analog", "analogy", "analytical", "analyze",
		"grape", "drape", "scrape", "rapeseed", "therapist", "therapy",
		"title", "titan", "titillate",
		"JavaScript", "CockpitDoor", "AssessmentReport",
	}
	for _, word := range safeWords {
		t.Run("safe: "+word, func(t *testing.T) { assertClean(t, en, word) })
	}
}

func TestEnglishMasking(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})
	result := en.Clean("what the fuck", nil)
	assertNotContains(t, result, "fuck")
	assertContains(t, result, "*")
}

func TestEnglishCleanText(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})
	assertClean(t, en, "hello world how are you")
}

func TestEnglishIsolation(t *testing.T) {
	en := mustNew(t, &terlik.Options{Language: "en"})
	assertClean(t, en, "siktir git")
	assertClean(t, en, "mierda")
	assertClean(t, en, "scheiße")
}
