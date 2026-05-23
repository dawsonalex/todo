# todo

![example workflow](https://github.com/dawsonalex/todo/workflows/Build/badge.svg)

A command-line tool for managing a [todo.txt](http://todotxt.org/) file.

## Usage

```
todo [flags] [item...]
```

Without positional arguments, `todo` lists the contents of your todo file. Pass
one or more positional arguments (or pipe lines via stdin) to add items.

### Flags

| Flag | Default | Description |
|------|---------|-------------|
| `-f <path>` | `~/todo.txt` | Path to the todo.txt file (overrides `TODO_FILE` env var) |
| `-s <field>` | `created` | Sort field: `priority`, `created`, or `completed` |
| `-q <term>` | | Filter term — repeatable, matched with AND logic (e.g. `-q @work -q +project`) |
| `-done` | | Include completed items in output |

### File resolution

The todo file is resolved in this order:

1. `-f` flag
2. `TODO_FILE` environment variable
3. `~/todo.txt`

### Examples

```sh
# List all incomplete items, sorted by creation date (default)
todo

# List items tagged @work, sorted by priority
todo -s priority -q @work

# Add an item from a positional argument
todo "(A) 2026-05-23 Fix the critical bug +work @laptop"

# Add items from a file
cat new-items.txt | todo

# Show completed items too
todo -done
```

## Development

### Prerequisites

- Go 1.25+

golangci-lint is declared as a `tool` dependency in `go.mod` and is fetched automatically by `go tool` — no separate install required.

### Common tasks

| Command | Description |
|---------|-------------|
| `make build` | Build the `todo` binary |
| `make run` | Build and run the binary |
| `make test` | Run tests with race detection and coverage |
| `make fmt` | Format Go code |
| `make vet` | Run `go vet` |
| `make lint` | Run `go tool golangci-lint` |
| `make commit-check` | Run all checks (fmt, vet, lint, test) — use before committing |
| `make clean` | Remove build artifacts |
| `make help` | List all available targets |
