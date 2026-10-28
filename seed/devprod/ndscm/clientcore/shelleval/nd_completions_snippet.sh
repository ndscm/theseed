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

  COMPREPLY=($(ndscm complete --cur="${cur}" --cword="${cword}" -- "${words[@]}" 2>/dev/null))
}

if [[ -n "${BASH_VERSION:-}" ]]; then
  complete -F _nd_completions nd
fi

if [[ -n "${ZSH_VERSION:-}" ]]; then
  autoload -Uz compinit && compinit 2>/dev/null
  autoload -Uz bashcompinit && bashcompinit 2>/dev/null
  complete -F _nd_completions nd
fi
