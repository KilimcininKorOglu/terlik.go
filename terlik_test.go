package terlik

// Internal function tests — these test unexported functions that can't be
// accessed from the external test package in tests/.
// All public API tests are in the tests/ directory.

import "testing"

func TestInternalCollapseRepeats(t *testing.T) {
	tests := []struct{ input, want string }{
		{"aaa", "a"},
		{"aa", "aa"},
		{"aaabbb", "ab"},
		{"hello", "hello"},
		{"siiiiik", "sik"},
	}
	for _, tt := range tests {
		got := collapseRepeats(tt.input)
		if got != tt.want {
			t.Errorf("collapseRepeats(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInternalExpandNumbers(t *testing.T) {
	expansions := [][2]string{
		{"100", "yuz"}, {"50", "elli"}, {"10", "on"}, {"2", "iki"},
	}
	tests := []struct{ input, want string }{
		{"s2k", "sikik"},
		{"a2b", "aikib"},
		{"2023 yilinda", "2023 yilinda"},
		{"8ok", "8ok"},
	}
	for _, tt := range tests {
		got := expandNumbers(tt.input, expansions)
		if got != tt.want {
			t.Errorf("expandNumbers(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInternalRemovePunctuation(t *testing.T) {
	tests := []struct{ input, want string }{
		{"s.i.k", "sik"},
		{"s-i-k", "sik"},
		{"hello! world", "hello! world"},
		{"test.", "test."},
	}
	for _, tt := range tests {
		got := removePunctuationBetweenLetters(tt.input)
		if got != tt.want {
			t.Errorf("removePunctuation(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInternalTurkishToLower(t *testing.T) {
	tests := []struct{ input, want string }{
		{"HELLO", "hello"},
		{"I", "ı"},
		{"İ", "i"},
	}
	for _, tt := range tests {
		got := turkishToLower(tt.input)
		if got != tt.want {
			t.Errorf("turkishToLower(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestInternalIsWordChar(t *testing.T) {
	if !isWordChar('a') {
		t.Error("'a' should be word char")
	}
	if !isWordChar('Z') {
		t.Error("'Z' should be word char")
	}
	if !isWordChar('5') {
		t.Error("'5' should be word char")
	}
	if isWordChar(' ') {
		t.Error("' ' should not be word char")
	}
	if isWordChar('$') {
		t.Error("'$' should not be word char")
	}
}
