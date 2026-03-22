package terlik_test

import (
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func TestApplyMask(t *testing.T) {
	t.Run("stars replaces with asterisks matching length", func(t *testing.T) {
		got := terlik.ApplyMask("siktir", terlik.MaskStars, "[***]")
		if got != "******" {
			t.Errorf("got %q, want %q", got, "******")
		}
	})
	t.Run("partial keeps first and last char", func(t *testing.T) {
		got := terlik.ApplyMask("siktir", terlik.MaskPartial, "[***]")
		if got != "s****r" {
			t.Errorf("got %q, want %q", got, "s****r")
		}
	})
	t.Run("partial handles short words", func(t *testing.T) {
		if got := terlik.ApplyMask("am", terlik.MaskPartial, ""); got != "**" {
			t.Errorf("am partial: got %q, want %q", got, "**")
		}
		if got := terlik.ApplyMask("a", terlik.MaskPartial, ""); got != "*" {
			t.Errorf("a partial: got %q, want %q", got, "*")
		}
	})
	t.Run("replace uses custom mask", func(t *testing.T) {
		if got := terlik.ApplyMask("siktir", terlik.MaskReplace, "[***]"); got != "[***]" {
			t.Errorf("got %q, want %q", got, "[***]")
		}
		if got := terlik.ApplyMask("siktir", terlik.MaskReplace, "***"); got != "***" {
			t.Errorf("got %q, want %q", got, "***")
		}
	})
}

func TestCleanText(t *testing.T) {
	matches := []terlik.MatchResult{
		{Word: "siktir", Root: "sik", Index: 7, Severity: terlik.SeverityHigh, Method: terlik.MethodPattern},
	}

	t.Run("replaces matched words with stars", func(t *testing.T) {
		result := terlik.CleanText("haydi, siktir git!", matches, terlik.MaskStars, "[***]")
		if result != "haydi, ****** git!" {
			t.Errorf("got %q, want %q", result, "haydi, ****** git!")
		}
	})
	t.Run("replaces with partial mask", func(t *testing.T) {
		result := terlik.CleanText("haydi, siktir git!", matches, terlik.MaskPartial, "[***]")
		if result != "haydi, s****r git!" {
			t.Errorf("got %q, want %q", result, "haydi, s****r git!")
		}
	})
	t.Run("replaces with custom mask", func(t *testing.T) {
		result := terlik.CleanText("haydi, siktir git!", matches, terlik.MaskReplace, "[küfür]")
		if result != "haydi, [küfür] git!" {
			t.Errorf("got %q, want %q", result, "haydi, [küfür] git!")
		}
	})
	t.Run("handles multiple matches", func(t *testing.T) {
		multi := []terlik.MatchResult{
			{Word: "siktir", Root: "sik", Index: 0, Severity: terlik.SeverityHigh, Method: terlik.MethodPattern},
			{Word: "aptal", Root: "aptal", Index: 11, Severity: terlik.SeverityLow, Method: terlik.MethodPattern},
		}
		result := terlik.CleanText("siktir lan aptal", multi, terlik.MaskStars, "[***]")
		if result != "****** lan *****" {
			t.Errorf("got %q, want %q", result, "****** lan *****")
		}
	})
	t.Run("returns original text when no matches", func(t *testing.T) {
		result := terlik.CleanText("merhaba dunya", nil, terlik.MaskStars, "[***]")
		if result != "merhaba dunya" {
			t.Errorf("got %q, want %q", result, "merhaba dunya")
		}
	})
}
