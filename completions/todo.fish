# Fish completion for the todo CLI.
#
# Installation:
#   todo completion fish > ~/.config/fish/completions/todo.fish
#
# Fish auto-loads files from that directory — no further setup needed.

# Disable default file completions for todo.
complete -c todo -f

# Helper: return the -f flag value if it has already been typed.
function __todo_file_flag
    set -l tokens (commandline -opc)
    set -l i 1
    while test $i -lt (count $tokens)
        if test $tokens[$i] = "-f"
            set -l j (math $i + 1)
            if test $j -le (count $tokens)
                echo $tokens[$j]
                return
            end
        end
        set i (math $i + 1)
    end
end

# Main completion function.
function __todo_completions
    set -l curword (commandline -t)

    # Only attempt tag completion when the current token contains @ or +.
    if not string match -q '*[@+]*' -- "$curword"
        return
    end

    # Resolve -f flag value, if present.
    set -l file_flag
    set -l fval (__todo_file_flag)
    if test -n "$fval"
        set file_flag -f $fval
    end

    # Ask the binary for completions.
    # Use the first token from the command line so this works for "./todo", full paths, etc.
    set -l cmd (commandline -opc)[1]
    set -l results ($cmd $file_flag --complete "$curword" 2>/dev/null)
    test (count $results) -eq 0; and return

    # Strip the rightmost sigil+partial to get the prefix.
    # e.g. "fix bug @wor" → prefix = "fix bug ", tag output = "@work"
    set -l prefix (string replace --regex '[@+][^@+]*$' '' -- "$curword")

    for tag in $results
        echo "$prefix$tag"
    end
end

complete -c todo -a '(__todo_completions)'
