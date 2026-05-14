function _nd_completions {
  local cur
  local prev
  local words
  local cword
  if type _init_completion &>/dev/null; then
    _init_completion || return
  else
    COMPREPLY=()
    cur="${COMP_WORDS[COMP_CWORD]}"
    prev="${COMP_WORDS[COMP_CWORD - 1]}"
    words=("${COMP_WORDS[@]}")
    cword=$COMP_CWORD
  fi

  if [[ $cword -eq 1 ]]; then
    COMPREPLY=($(compgen -W "apply bootstrap build change check connect cut dev format lock run setup shell submit sync test tidy uncut vendor" -- "$cur"))
    return
  fi

  # We keep each command separate to allow for more specific completions in the future
  case "${words[1]}" in
  "change")
    local branches=$(git branch --list 'change/*' 2>/dev/null | sed 's/^[* ]*//' | sed 's|^change/||')
    COMPREPLY=($(compgen -W "$branches" -- "$cur"))
    ;;
  "submit")
    local branches=$(git branch --list 'change/*' 2>/dev/null | sed 's/^[* ]*//' | sed 's|^change/||')
    COMPREPLY=($(compgen -W "$branches" -- "$cur"))
    ;;
  esac
}

if [[ -n "${BASH_VERSION:-}" ]]; then
  complete -F _nd_completions nd
fi

if [[ -n "${ZSH_VERSION:-}" ]]; then
  autoload -Uz compinit && compinit 2>/dev/null
  autoload -Uz bashcompinit && bashcompinit 2>/dev/null
  complete -F _nd_completions nd
fi
