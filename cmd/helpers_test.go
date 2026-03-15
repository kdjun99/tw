package cmd

import (
	"os"
	"path/filepath"
	"testing"
)

func TestShortBranch(t *testing.T) {
	tests := []struct {
		input string
		want  string
	}{
		{"feature/login", "login"},
		{"fix/bug-123", "bug-123"},
		{"main", "main"},
		{"feature/nested/deep", "deep"},
		{"", ""},
	}
	for _, tt := range tests {
		got := shortBranch(tt.input)
		if got != tt.want {
			t.Errorf("shortBranch(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}

func TestSplitLast(t *testing.T) {
	tests := []struct {
		s    string
		sep  string
		want []string
	}{
		{"feature/login", "/", []string{"feature", "login"}},
		{"a/b/c", "/", []string{"a/b", "c"}},
		{"no-slash", "/", []string{"no-slash"}},
		{"", "/", []string{""}},
	}
	for _, tt := range tests {
		got := splitLast(tt.s, tt.sep)
		if len(got) != len(tt.want) {
			t.Errorf("splitLast(%q, %q) = %v, want %v", tt.s, tt.sep, got, tt.want)
			continue
		}
		for i := range got {
			if got[i] != tt.want[i] {
				t.Errorf("splitLast(%q, %q)[%d] = %q, want %q", tt.s, tt.sep, i, got[i], tt.want[i])
			}
		}
	}
}

func TestShortenPath(t *testing.T) {
	home, _ := os.UserHomeDir()

	tests := []struct {
		input string
		want  string
	}{
		{filepath.Join(home, "dev/myapp"), "~/dev/myapp"},
		{"/usr/local/bin", "/usr/local/bin"},
		{home, "~"},
	}
	for _, tt := range tests {
		got := shortenPath(tt.input)
		if got != tt.want {
			t.Errorf("shortenPath(%q) = %q, want %q", tt.input, got, tt.want)
		}
	}
}
