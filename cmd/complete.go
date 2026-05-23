package main

import (
	"bufio"
	"fmt"
	"os"
	"sort"
	"strings"

	"github.com/dawsonalex/todo"
	"github.com/dawsonalex/todo/completions"
)

// runCompletion handles the "todo completion <shell>" subcommand.
// It writes the embedded completion script for the named shell to stdout.
func runCompletion(args []string) {
	if len(args) != 1 {
		fmt.Fprintln(os.Stderr, "usage: todo completion [bash|fish|zsh]")
		os.Exit(1)
	}

	var name string
	switch args[0] {
	case "zsh":
		name = "todo.zsh"
	case "bash":
		name = "todo.bash"
	case "fish":
		name = "todo.fish"
	default:
		fmt.Fprintf(os.Stderr, "todo: unknown shell %q (want bash, fish, or zsh)\n", args[0])
		os.Exit(1)
	}

	data, err := completions.FS.ReadFile(name)
	if err != nil {
		fmt.Fprintf(os.Stderr, "todo: reading completion script: %v\n", err)
		os.Exit(1)
	}
	_, _ = os.Stdout.Write(data)
}

// handleCompletion writes tab-completion candidates for word to stdout and exits.
// word is the raw value of the --complete flag; it may be a bare tag token
// (e.g. "@wor") or a partial string containing a tag (e.g. "fix bug @wor").
// path is the resolved todo file path.
func handleCompletion(word, path string) {
	sigil, partial := extractTag(word)
	if sigil == 0 {
		os.Exit(0)
	}

	list, err := todo.ReadFile(path)
	if err != nil {
		os.Exit(0)
	}

	tags := collectTags(list.GetAll(), sigil)
	w := bufio.NewWriter(os.Stdout)
	for _, tag := range tags {
		if fuzzyMatch(partial, tag) {
			_, _ = fmt.Fprintf(w, "%c%s\n", sigil, tag)
		}
	}
	_ = w.Flush()
	os.Exit(0)
}

// extractTag scans the space-delimited tokens of word right-to-left and returns
// the sigil ('@' or '+') and partial value of the rightmost tag token found.
// Returns sigil=0 if no tag token is found.
func extractTag(word string) (sigil byte, partial string) {
	fields := strings.Fields(word)
	for i := len(fields) - 1; i >= 0; i-- {
		f := fields[i]
		if len(f) > 0 && (f[0] == '@' || f[0] == '+') {
			return f[0], f[1:]
		}
	}
	return 0, ""
}

// fuzzyMatch reports whether partial is a case-insensitive subsequence of candidate.
// An empty partial always matches.
func fuzzyMatch(partial, candidate string) bool {
	if partial == "" {
		return true
	}
	p := strings.ToLower(partial)
	c := strings.ToLower(candidate)
	pi := 0
	for ci := 0; ci < len(c) && pi < len(p); ci++ {
		if c[ci] == p[pi] {
			pi++
		}
	}
	return pi == len(p)
}

// collectTags returns a sorted, deduplicated slice of all context ('@') or
// project ('+') values found in items, depending on sigil.
func collectTags(items []todo.Item, sigil byte) []string {
	seen := make(map[string]struct{})
	var tags []string
	for _, item := range items {
		var source []string
		if sigil == '@' {
			source = item.Contexts
		} else {
			source = item.Projects
		}
		for _, t := range source {
			if _, ok := seen[t]; !ok {
				seen[t] = struct{}{}
				tags = append(tags, t)
			}
		}
	}
	sort.Strings(tags)
	return tags
}
