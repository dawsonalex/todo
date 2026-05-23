# Bash completion for the todo CLI.
#
# Installation:
#   todo completion bash > ~/.bash_completion.d/todo
#   source ~/.bash_completion.d/todo
#
# Or for a single session:
#   source <(todo completion bash)

_todo_complete() {
    local curword="${COMP_WORDS[COMP_CWORD]}"

    # Respect -f flag if the user has already typed it.
    local file_flag=()
    local i
    for (( i = 1; i < COMP_CWORD; i++ )); do
        if [[ "${COMP_WORDS[i]}" == "-f" ]] && (( i + 1 < COMP_CWORD )); then
            file_flag=("-f" "${COMP_WORDS[i+1]}")
            break
        fi
    done

    # Only attempt tag completion when the current word contains @ or +.
    if [[ "$curword" != *@* && "$curword" != *+* ]]; then
        return
    fi

    # Get completions from the binary.
    # Use $COMP_WORDS[0] so this works whether the command is "todo", "./todo", or a full path.
    local raw
    raw=$("${COMP_WORDS[0]}" "${file_flag[@]}" --complete "$curword" 2>/dev/null)
    [[ -z "$raw" ]] && return

    # Find the rightmost sigil by comparing suffix lengths.
    local after_at="${curword##*@}"
    local after_plus="${curword##*+}"
    local last_tag
    if (( ${#after_at} <= ${#after_plus} )); then
        last_tag="@${after_at}"
    else
        last_tag="+${after_plus}"
    fi

    # prefix is everything before the rightmost sigil+partial.
    local prefix="${curword%${last_tag}}"

    # Build COMPREPLY with full replacement words.
    COMPREPLY=()
    local tag
    while IFS= read -r tag; do
        [[ -n "$tag" ]] && COMPREPLY+=("${prefix}${tag}")
    done <<< "$raw"
}

# -o nospace: suppress the trailing space so users can keep typing.
complete -o nospace -F _todo_complete todo
