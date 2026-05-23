#compdef todo

# Zsh completion for the todo CLI.
#
# Installation:
#   mkdir -p ~/.zsh/completions
#   todo completion zsh > ~/.zsh/completions/_todo
#
#   In ~/.zshrc, before compinit:
#     fpath=(~/.zsh/completions $fpath)
#
#   Then reload: exec zsh

_todo() {
    local curword="${words[CURRENT]}"

    # Respect -f flag if the user has already typed it.
    local file_flag=()
    local i
    for (( i = 1; i < CURRENT; i++ )); do
        if [[ "${words[i]}" == "-f" ]] && (( i + 1 < CURRENT )); then
            file_flag=("-f" "${words[i+1]}")
            break
        fi
    done

    # Only attempt tag completion when the current word contains @ or +.
    if [[ "$curword" != *@* && "$curword" != *+* ]]; then
        return
    fi

    # Ask the binary for completions, passing the full current word.
    # The binary handles both bare "@wor" and in-string "fix bug @wor" cases.
    # Use $words[1] so this works whether the command is "todo", "./todo", or a full path.
    local completions
    completions=("${(@f)$("${words[1]}" "${file_flag[@]}" --complete "$curword" 2>/dev/null)}")
    (( ${#completions[@]} == 0 )) && return

    # Find the rightmost sigil in curword by comparing which suffix is shorter:
    # the part after the last @ vs the part after the last +.
    local after_at="${curword##*@}"
    local after_plus="${curword##*+}"
    local last_tag
    if (( ${#after_at} <= ${#after_plus} )); then
        last_tag="@${after_at}"
    else
        last_tag="+${after_plus}"
    fi

    # prefix is everything before the last sigil+partial, e.g. "fix bug ".
    local prefix="${curword%${last_tag}}"

    # Build full replacement words and hand them to zsh.
    local -a display
    local tag
    for tag in "${completions[@]}"; do
        display+=("${prefix}${tag}")
    done

    # -Q: don't re-quote; -S '': no trailing space (user continues typing).
    compadd -Q -S '' -a display
}

# Register the completion function explicitly so both installation methods work:
#   source <(todo completion zsh)   — compdef is processed here
#   autoload via $fpath             — #compdef at the top handles it
compdef _todo todo

_todo "$@"
