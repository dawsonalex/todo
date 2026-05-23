package main

import (
	"bufio"
	"errors"
	"flag"
	"fmt"
	"io"
	"os"
	"os/user"
	"path/filepath"
	"sort"
	"strings"
	"time"

	"github.com/dawsonalex/todo"
)

// queryFlag is a repeatable -q flag value.
type queryFlag []string

func (q *queryFlag) String() string { return strings.Join(*q, ", ") }
func (q *queryFlag) Set(s string) error {
	*q = append(*q, s)
	return nil
}

func main() {
	// Resolve piped stdin here so run() receives nil when stdin is a terminal.
	var stdin io.Reader
	if isStdinPiped() {
		stdin = os.Stdin
	}
	os.Exit(run(os.Args[1:], stdin, os.Stdout, os.Stderr))
}

func run(args []string, stdin io.Reader, stdout, stderr io.Writer) int {
	if len(args) > 0 && args[0] == "completion" {
		runCompletion(args[1:])
		return 0
	}

	fs := flag.NewFlagSet("todo", flag.ContinueOnError)
	fs.SetOutput(stderr)
	fs.Usage = func() {
		_, _ = fmt.Fprintf(stderr, "Usage:\n  todo [flags] [item...]\n  todo completion <shell>\n\nSubcommands:\n  completion <shell>   print the tab-completion script for bash, fish, or zsh\n\nFlags:\n")
		fs.PrintDefaults()
	}

	var queries queryFlag
	sortField := fs.String("s", "created", "sort field: priority, created, completed")
	filePath := fs.String("f", "", "path to todo.txt file (overrides TODO_FILE env var)")
	showDone := fs.Bool("done", false, "include completed items in output")
	verbose := fs.Bool("v", false, "print the resolved todo.txt path")
	completeWord := fs.String("complete", "", "output tab completions for word (used by shell completion scripts)")
	fs.Var(&queries, "q", "filter term, repeatable with AND logic (e.g. -q @work -q +project)")

	if err := fs.Parse(args); err != nil {
		if errors.Is(err, flag.ErrHelp) {
			return 0
		}
		return 1
	}

	path, err := resolvePath(*filePath)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "todo: resolving path: %v\n", err)
		return 1
	}

	if *completeWord != "" {
		handleCompletion(*completeWord, path)
		return 0
	}

	if *verbose {
		_, _ = fmt.Fprintf(stdout, "todo file: %s\n", path)
	}

	list, err := todo.ReadFile(path)
	if err != nil {
		_, _ = fmt.Fprintf(stderr, "todo: reading %s: %v\n", path, err)
		return 1
	}

	adding := false

	if stdin != nil {
		if err := addFromReader(list, stdin); err != nil {
			_, _ = fmt.Fprintf(stderr, "todo: reading stdin: %v\n", err)
			return 1
		}
		adding = true
	}

	if posArgs := fs.Args(); len(posArgs) > 0 {
		text := strings.Join(posArgs, " ")
		if err := addItem(list, text); err != nil {
			_, _ = fmt.Fprintf(stderr, "todo: parsing item %q: %v\n", text, err)
			return 1
		}
		adding = true
	}

	if adding {
		if err := todo.WriteFile(path, list); err != nil {
			_, _ = fmt.Fprintf(stderr, "todo: writing %s: %v\n", path, err)
			return 1
		}
		return 0
	}

	// List mode: filter, sort, print.
	items := list.GetAll()
	items = filterItems(items, queries, *showDone)
	items = sortItems(items, *sortField)
	printItems(items, stdout)
	return 0
}

// resolvePath returns the todo.txt path to use: -f flag > TODO_FILE env > ~/todo.txt.
func resolvePath(flagVal string) (string, error) {
	if flagVal != "" {
		return flagVal, nil
	}
	if env, ok := os.LookupEnv("TODO_FILE"); ok && env != "" {
		return env, nil
	}
	u, err := user.Current()
	if err != nil {
		return "", fmt.Errorf("looking up home directory: %w", err)
	}
	return filepath.Join(u.HomeDir, "todo.txt"), nil
}

// isStdinPiped reports whether stdin is a pipe rather than a terminal.
func isStdinPiped() bool {
	stat, err := os.Stdin.Stat()
	if err != nil {
		return false
	}
	return (stat.Mode() & os.ModeCharDevice) == 0
}

// addFromReader reads newline-delimited todo.txt lines from r and adds them to the list.
func addFromReader(list *todo.List, r io.Reader) error {
	scanner := bufio.NewScanner(r)
	for scanner.Scan() {
		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}
		if err := addItem(list, line); err != nil {
			return err
		}
	}
	return scanner.Err()
}

// addItem parses a todo.txt line and appends it to the list.
// If no creation date is present in the text, today's date is set.
func addItem(list *todo.List, text string) error {
	var item todo.Item
	if err := item.UnmarshalText([]byte(text)); err != nil {
		return err
	}
	if item.CreatedDate.IsZero() {
		item.CreatedDate = time.Now().Truncate(24 * time.Hour)
	}
	list.Add(item)
	return nil
}

// filterItems returns items matching all query terms and respecting the showDone flag.
// Matching is a case-sensitive substring check against the todo.txt representation of each item.
func filterItems(items []todo.Item, queries []string, showDone bool) []todo.Item {
	out := items[:0:0]
	for _, item := range items {
		if item.Done && !showDone {
			continue
		}
		text, _ := item.MarshalText()
		line := string(text)
		match := true
		for _, q := range queries {
			if !strings.Contains(line, q) {
				match = false
				break
			}
		}
		if match {
			out = append(out, item)
		}
	}
	return out
}

// sortItems sorts items by the named field. Unknown fields leave order unchanged.
func sortItems(items []todo.Item, field string) []todo.Item {
	switch field {
	case "priority":
		sort.SliceStable(items, func(i, j int) bool {
			pi, pj := items[i].Priority, items[j].Priority
			if pi.Valid() && pj.Valid() {
				return pi < pj
			}
			return pi.Valid() // valid priorities sort before invalid (no priority)
		})
	case "completed":
		sort.SliceStable(items, func(i, j int) bool {
			ti, tj := items[i].CompletedDate, items[j].CompletedDate
			if ti.IsZero() != tj.IsZero() {
				return !ti.IsZero() // items with a completed date sort before those without
			}
			return ti.Before(tj)
		})
	default: // "created"
		sort.SliceStable(items, func(i, j int) bool {
			ti, tj := items[i].CreatedDate, items[j].CreatedDate
			if ti.IsZero() != tj.IsZero() {
				return !ti.IsZero()
			}
			return ti.Before(tj)
		})
	}
	return items
}

func printItems(items []todo.Item, w io.Writer) {
	bw := bufio.NewWriter(w)
	for _, item := range items {
		text, _ := item.MarshalText()
		_, _ = fmt.Fprintf(bw, "%s\n", text)
	}
	_ = bw.Flush()
}
