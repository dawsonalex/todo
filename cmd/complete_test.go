package main

import (
	"testing"

	"github.com/dawsonalex/todo"
)

func TestExtractTag(t *testing.T) {
	tests := []struct {
		name        string
		word        string
		wantSigil   byte
		wantPartial string
	}{
		{"bare context", "@wor", '@', "wor"},
		{"bare project", "+pr", '+', "pr"},
		{"sigil only context", "@", '@', ""},
		{"sigil only project", "+", '+', ""},
		{"in-string context", "fix bug @wor", '@', "wor"},
		{"in-string project", "fix bug +pr", '+', "pr"},
		{"rightmost wins context", "@work +pr", '+', "pr"},
		{"rightmost wins project", "+proj @wor", '@', "wor"},
		{"no tag", "fix bug", 0, ""},
		{"empty string", "", 0, ""},
	}
	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			gotSigil, gotPartial := extractTag(tc.word)
			if gotSigil != tc.wantSigil || gotPartial != tc.wantPartial {
				t.Errorf("extractTag(%q) = (%q, %q), want (%q, %q)",
					tc.word, gotSigil, gotPartial, tc.wantSigil, tc.wantPartial)
			}
		})
	}
}

func TestFuzzyMatch(t *testing.T) {
	tests := []struct {
		partial   string
		candidate string
		want      bool
	}{
		{"", "anything", true},
		{"", "", true},
		{"work", "work", true},
		{"wor", "work", true},
		{"wk", "work", true}, // subsequence: w...k
		{"WK", "work", true}, // case-insensitive
		{"wk", "WORK", true}, // case-insensitive candidate
		{"xyz", "work", false},
		{"wok", "work", true}, // subsequence: w(0) o(1) k(3) — r is skipped
		{"wr", "work", true},  // w...r subsequence
		{"z", "work", false},
	}
	for _, tc := range tests {
		t.Run(tc.partial+"→"+tc.candidate, func(t *testing.T) {
			got := fuzzyMatch(tc.partial, tc.candidate)
			if got != tc.want {
				t.Errorf("fuzzyMatch(%q, %q) = %v, want %v", tc.partial, tc.candidate, got, tc.want)
			}
		})
	}
}

func TestCollectTags(t *testing.T) {
	items := []todo.Item{
		{Contexts: []string{"work", "home"}, Projects: []string{"laptop"}},
		{Contexts: []string{"work", "weekend"}, Projects: []string{"laptop", "mobile"}},
		{Contexts: []string{}, Projects: []string{}},
	}

	t.Run("contexts sorted and deduped", func(t *testing.T) {
		got := collectTags(items, '@')
		want := []string{"home", "weekend", "work"}
		if !sliceEqual(got, want) {
			t.Errorf("collectTags contexts = %v, want %v", got, want)
		}
	})

	t.Run("projects sorted and deduped", func(t *testing.T) {
		got := collectTags(items, '+')
		want := []string{"laptop", "mobile"}
		if !sliceEqual(got, want) {
			t.Errorf("collectTags projects = %v, want %v", got, want)
		}
	})

	t.Run("empty items", func(t *testing.T) {
		got := collectTags(nil, '@')
		if len(got) != 0 {
			t.Errorf("collectTags(nil) = %v, want empty", got)
		}
	})
}

func sliceEqual(a, b []string) bool {
	if len(a) != len(b) {
		return false
	}
	for i := range a {
		if a[i] != b[i] {
			return false
		}
	}
	return true
}
