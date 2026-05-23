package completions

import "embed"

//go:embed todo.zsh todo.bash todo.fish
var FS embed.FS
