package terlik_test

import (
	"strings"
	"github.com/KilimcininKorOglu/terlik.go"
	"testing"
)

func mustNew(t *testing.T, opts *terlik.Options) *terlik.Terlik {
	t.Helper()
	instance, err := terlik.New(opts)
	if err != nil {
		t.Fatalf("terlik.New failed: %v", err)
	}
	return instance
}

func assertDetects(t *testing.T, instance *terlik.Terlik, text string) {
	t.Helper()
	if !instance.ContainsProfanity(text, nil) {
		t.Errorf("expected %q to be detected as profanity", text)
	}
}

func assertClean(t *testing.T, instance *terlik.Terlik, text string) {
	t.Helper()
	if instance.ContainsProfanity(text, nil) {
		t.Errorf("expected %q to be clean (not profanity)", text)
	}
}

func assertDetectsRoot(t *testing.T, instance *terlik.Terlik, text, expectedRoot string) {
	t.Helper()
	matches := instance.GetMatches(text, nil)
	if len(matches) == 0 {
		t.Errorf("expected %q to be detected", text)
		return
	}
	for _, m := range matches {
		if m.Root == expectedRoot {
			return
		}
	}
	var roots []string
	for _, m := range matches {
		roots = append(roots, m.Root)
	}
	t.Errorf("expected %q to match root %q, got: %s", text, expectedRoot, strings.Join(roots, ", "))
}

func assertContains(t *testing.T, s, substr string) {
	t.Helper()
	if !strings.Contains(s, substr) {
		t.Errorf("expected %q to contain %q", s, substr)
	}
}

func assertNotContains(t *testing.T, s, substr string) {
	t.Helper()
	if strings.Contains(s, substr) {
		t.Errorf("expected %q to NOT contain %q", s, substr)
	}
}
