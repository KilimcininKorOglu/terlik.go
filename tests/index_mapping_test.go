package terlik_test

import (
	"testing"

	terlik "github.com/KilimcininKorOglu/terlik.go"
)

// TestMatchIndexAccuracy validates that MatchResult.Index values are correct
// byte offsets into the original text. This directly exercises the
// mapNormalizedToOriginal code path which has been the root cause of
// BUG-003, BUG-009, and BUG-010.
func TestMatchIndexAccuracy(t *testing.T) {
	tests := []struct {
		name     string
		lang     string
		text     string
		wantRoot string
		wantIdx  int
	}{
		{
			name:     "Turkish uppercase I multi-word",
			lang:     "tr",
			text:     "BU SİKTİR",
			wantRoot: "sik",
			wantIdx:  3, // "SİKTİR" starts at byte 3
		},
		{
			name:     "Turkish plain lowercase",
			lang:     "tr",
			text:     "bu siktir",
			wantRoot: "sik",
			wantIdx:  3,
		},
		{
			name:     "multi-byte chars before profanity",
			lang:     "tr",
			text:     "şeker siktir",
			wantRoot: "sik",
			wantIdx:  7, // ş=2 + e=1 + k=1 + e=1 + r=1 + space=1 = 7
		},
		{
			name:     "profanity at start",
			lang:     "tr",
			text:     "siktir git",
			wantRoot: "sik",
			wantIdx:  0,
		},
		{
			name:     "English simple",
			lang:     "en",
			text:     "what the fuck",
			wantRoot: "fuck",
			wantIdx:  9,
		},
		{
			name:     "English profanity at start",
			lang:     "en",
			text:     "fuck off",
			wantRoot: "fuck",
			wantIdx:  0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			instance := mustNew(t, &terlik.Options{Language: tt.lang})
			matches := instance.GetMatches(tt.text, nil)
			if len(matches) == 0 {
				t.Fatalf("expected %q to be detected", tt.text)
			}
			found := false
			for _, m := range matches {
				if m.Root == tt.wantRoot && m.Index == tt.wantIdx {
					found = true
					break
				}
			}
			if !found {
				for _, m := range matches {
					t.Logf("  match: Root=%q Index=%d Word=%q", m.Root, m.Index, m.Word)
				}
				t.Errorf("expected match with Root=%q Index=%d, not found in results", tt.wantRoot, tt.wantIdx)
			}
		})
	}
}

func TestMatchIndexWithLeadingWhitespace(t *testing.T) {
	instance := mustNew(t, &terlik.Options{Language: "tr"})

	// Leading whitespace should not shift reported Index
	text := "  siktir git"
	matches := instance.GetMatches(text, nil)
	if len(matches) == 0 {
		t.Fatalf("expected %q to be detected", text)
	}
	// "siktir" starts at byte 2 (two spaces)
	for _, m := range matches {
		if m.Root == "sik" {
			if m.Index != 2 {
				t.Errorf("expected Index=2 for leading-whitespace text, got %d", m.Index)
			}
			return
		}
	}
	t.Error("expected match with root 'sik'")
}

func TestMatchIndexWithInvisibleChars(t *testing.T) {
	instance := mustNew(t, &terlik.Options{Language: "tr"})

	// Invisible chars between words should not break offset tracking
	text := "test \u200B siktir"
	matches := instance.GetMatches(text, nil)
	if len(matches) == 0 {
		t.Fatalf("expected %q to be detected", text)
	}
	for _, m := range matches {
		if m.Root == "sik" {
			// \u200B is 3 bytes. "test"=4 + " "=1 + \u200B=3 + " "=1 = 9. "siktir" at byte 9.
			if m.Index != 9 {
				t.Errorf("expected Index=9 for invisible-char text, got %d (Word=%q)", m.Index, m.Word)
			}
			return
		}
	}
	t.Error("expected match with root 'sik'")
}
