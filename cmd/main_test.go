package main

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/dawsonalex/todo"
)

// writeRawFile writes a raw todo.txt string to a temp file and returns the path.
func writeRawFile(t *testing.T, content string) string {
	t.Helper()
	dir := t.TempDir()
	path := filepath.Join(dir, "todo.txt")
	if err := os.WriteFile(path, []byte(content), 0600); err != nil {
		t.Fatalf("writeRawFile: %v", err)
	}
	return path
}

// emptyFilePath returns the path to a non-existent temp todo.txt (ReadFile treats these as empty).
func emptyFilePath(t *testing.T) string {
	t.Helper()
	return filepath.Join(t.TempDir(), "todo.txt")
}

// readItemsFromFile reads all items from a todo.txt file.
func readItemsFromFile(t *testing.T, path string) []todo.Item {
	t.Helper()
	list, err := todo.ReadFile(path)
	if err != nil {
		t.Fatalf("readItemsFromFile: %v", err)
	}
	return list.GetAll()
}

// outputLines splits a string on newlines, discarding empty lines.
func outputLines(s string) []string {
	var lines []string
	for _, line := range strings.Split(strings.TrimSpace(s), "\n") {
		if line != "" {
			lines = append(lines, line)
		}
	}
	return lines
}

// TestRun_AddMultiWordArg is the regression test for the bug where
// `./todo one new message` created three items instead of one.
func TestRun_AddMultiWordArg(t *testing.T) {
	path := emptyFilePath(t)
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path, "one", "new", "message"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}

	items := readItemsFromFile(t, path)
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	if got := items[0].Message; got != "one new message" {
		t.Errorf("message = %q, want %q", got, "one new message")
	}
}

func TestRun_AddSingleWordArg(t *testing.T) {
	path := emptyFilePath(t)
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path, "task"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}

	items := readItemsFromFile(t, path)
	if len(items) != 1 {
		t.Fatalf("want 1 item, got %d", len(items))
	}
	if got := items[0].Message; got != "task" {
		t.Errorf("message = %q, want %q", got, "task")
	}
}

func TestRun_AddFromStdin(t *testing.T) {
	path := emptyFilePath(t)
	stdin := strings.NewReader("buy milk\ncall dentist\n")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path}, stdin, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}

	items := readItemsFromFile(t, path)
	if len(items) != 2 {
		t.Fatalf("want 2 items, got %d", len(items))
	}
	if items[0].Message != "buy milk" {
		t.Errorf("item 0 message = %q, want %q", items[0].Message, "buy milk")
	}
	if items[1].Message != "call dentist" {
		t.Errorf("item 1 message = %q, want %q", items[1].Message, "call dentist")
	}
}

func TestRun_QueryFilter(t *testing.T) {
	path := writeRawFile(t, "fix bug @work\nbuy milk @home\nwrite tests @work\n")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path, "-q", "@work"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}

	lines := outputLines(stdout.String())
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %v", len(lines), lines)
	}
	for _, line := range lines {
		if !strings.Contains(line, "@work") {
			t.Errorf("line %q does not contain @work", line)
		}
	}
}

func TestRun_QueryFilterAND(t *testing.T) {
	path := writeRawFile(t, "fix bug @work +backend\nbuy milk @home\nwrite tests @work +frontend\n")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path, "-q", "@work", "-q", "+backend"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}

	lines := outputLines(stdout.String())
	if len(lines) != 1 {
		t.Fatalf("want 1 line, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "+backend") {
		t.Errorf("line %q does not contain +backend", lines[0])
	}
}

func TestRun_ShowDoneFlag(t *testing.T) {
	// "x 2024-01-01 done task" → Done=true, CreatedDate=2024-01-01, Message="done task"
	path := writeRawFile(t, "x 2024-01-01 done task\nactive task\n")

	t.Run("done items excluded by default", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := run([]string{"-f", path}, nil, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("run exited %d: %s", code, stderr.String())
		}
		lines := outputLines(stdout.String())
		if len(lines) != 1 {
			t.Fatalf("want 1 line, got %d: %v", len(lines), lines)
		}
		if strings.Contains(lines[0], "done task") {
			t.Errorf("done item should be excluded, got %q", lines[0])
		}
	})

	t.Run("done items included with -done", func(t *testing.T) {
		var stdout, stderr bytes.Buffer
		code := run([]string{"-f", path, "-done"}, nil, &stdout, &stderr)
		if code != 0 {
			t.Fatalf("run exited %d: %s", code, stderr.String())
		}
		lines := outputLines(stdout.String())
		if len(lines) != 2 {
			t.Fatalf("want 2 lines, got %d: %v", len(lines), lines)
		}
	})
}

func TestRun_SortByPriority(t *testing.T) {
	path := writeRawFile(t, "(B) second priority\n(A) first priority\nno priority task\n")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path, "-s", "priority"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}

	lines := outputLines(stdout.String())
	if len(lines) != 3 {
		t.Fatalf("want 3 lines, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "(A)") {
		t.Errorf("line 0 = %q, want (A) item first", lines[0])
	}
	if !strings.Contains(lines[1], "(B)") {
		t.Errorf("line 1 = %q, want (B) item second", lines[1])
	}
	if strings.Contains(lines[2], "(") {
		t.Errorf("line 2 = %q, want no-priority item last", lines[2])
	}
}

func TestRun_SortByCreated(t *testing.T) {
	path := writeRawFile(t, "2024-01-02 second\n2024-01-01 first\n")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path, "-s", "created"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}

	lines := outputLines(stdout.String())
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "first") {
		t.Errorf("line 0 = %q, want earlier-dated item first", lines[0])
	}
	if !strings.Contains(lines[1], "second") {
		t.Errorf("line 1 = %q, want later-dated item second", lines[1])
	}
}

func TestRun_SortByCompleted(t *testing.T) {
	// Format: "x <completed-date> <created-date> <message>"
	path := writeRawFile(t, "x 2024-01-03 2024-01-01 done later\nx 2024-01-02 2024-01-01 done earlier\n")
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path, "-s", "completed", "-done"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}

	lines := outputLines(stdout.String())
	if len(lines) != 2 {
		t.Fatalf("want 2 lines, got %d: %v", len(lines), lines)
	}
	if !strings.Contains(lines[0], "done earlier") {
		t.Errorf("line 0 = %q, want earlier-completed item first", lines[0])
	}
}

func TestRun_VerboseFlag(t *testing.T) {
	path := emptyFilePath(t)
	var stdout, stderr bytes.Buffer

	code := run([]string{"-f", path, "-v"}, nil, &stdout, &stderr)
	if code != 0 {
		t.Fatalf("run exited %d: %s", code, stderr.String())
	}
	if !strings.Contains(stdout.String(), path) {
		t.Errorf("stdout %q should contain resolved path %q", stdout.String(), path)
	}
}
