package main

import (
	"bufio"
	"flag"
	"fmt"
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
	if len(os.Args) > 1 && os.Args[1] == "completion" {
		runCompletion(os.Args[2:])
		return
	}

	flag.Usage = func() {
		_, _ = fmt.Fprintf(flag.CommandLine.Output(), "Usage:\n  todo [flags] [item...]\n  todo completion <shell>\n\nSubcommands:\n  completion <shell>   print the tab-completion script for bash, fish, or zsh\n\nFlags:\n")
		flag.PrintDefaults()
	}

	var queries queryFlag
	sortField := flag.String("s", "created", "sort field: priority, created, completed")
	filePath := flag.String("f", "", "path to todo.txt file (overrides TODO_FILE env var)")
	showDone := flag.Bool("done", false, "include completed items in output")
	verbose := flag.Bool("v", false, "print the resolved todo.txt path")
	completeWord := flag.String("complete", "", "output tab completions for word (used by shell completion scripts)")
	flag.Var(&queries, "q", "filter term, repeatable with AND logic (e.g. -q @work -q +project)")
	flag.Parse()

	path, err := resolvePath(*filePath)
	if err != nil {
		fatalf("resolving path: %v", err)
	}

	if *completeWord != "" {
		handleCompletion(*completeWord, path)
	}

	if *verbose {
		fmt.Printf("todo file: %s\n", path)
	}

	list, err := todo.ReadFile(path)
	if err != nil {
		fatalf("reading %s: %v", path, err)
	}

	// Determine whether we are in add mode (stdin pipe or positional args).
	adding := false

	if isStdinPiped() {
		if err := addFromReader(list, os.Stdin); err != nil {
			fatalf("reading stdin: %v", err)
		}
		adding = true
	}

	if args := flag.Args(); len(args) > 0 {
		for _, arg := range args {
			if err := addItem(list, arg); err != nil {
				fatalf("parsing item %q: %v", arg, err)
			}
		}
		adding = true
	}

	if adding {
		if err := todo.WriteFile(path, list); err != nil {
			fatalf("writing %s: %v", path, err)
		}
		return
	}

	// List mode: filter, sort, print.
	items := list.GetAll()
	items = filterItems(items, queries, *showDone)
	items = sortItems(items, *sortField)
	printItems(items)
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
func addFromReader(list *todo.List, r *os.File) error {
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

func printItems(items []todo.Item) {
	w := bufio.NewWriter(os.Stdout)
	for _, item := range items {
		text, _ := item.MarshalText()
		_, _ = fmt.Fprintf(w, "%s\n", text)
	}
	_ = w.Flush()
}

func fatalf(format string, args ...any) {
	_, _ = fmt.Fprintf(os.Stderr, "todo: "+format+"\n", args...)
	os.Exit(1)
}
