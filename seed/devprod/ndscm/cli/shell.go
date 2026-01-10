package main

import (
	"fmt"
	"log"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
)

func NdShell(args []string, ndConfig *common.NdConfig) error {
	if !ndConfig.ShellEval {
		log.Printf("\x1b[33mWarning: It's recommended to run nd-shell with --shell-eval\x1b[0m")
	}
	if len(args) != 1 {
		return common.WrapTrace(fmt.Errorf("nd-shell usage: eval \"$(ndscm --shell-eval shell)\""))
	}
	shellEval := `
function nd {
  if [[ "$#" -eq 0 ]]; then
    ndscm
  elif [[ "$#" -ge 1 ]]; then
    case "${1}" in
    "dev")
      eval "$(ndscm --shell-eval ${1} ${@:2})"
      ;;
    *)
      ndscm ${@:1}
      ;;
    esac
  fi
}

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
    COMPREPLY=($(compgen -W "cut dev review setup shell sync" -- "$cur"))
    return
  fi

  case "${words[1]}" in
  review)
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
`
	if ndConfig.Dry {
		log.Printf("Shell eval: %v", shellEval)
		return nil
	}
	if ndConfig.ShellEval {
		fmt.Printf("%v", shellEval)
	}
	return nil
}
