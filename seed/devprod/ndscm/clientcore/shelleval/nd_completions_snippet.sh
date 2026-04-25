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
    COMPREPLY=($(compgen -W "cut dev setup shell submit sync" -- "$cur"))
    return
  fi

  case "${words[1]}" in
  submit)
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
