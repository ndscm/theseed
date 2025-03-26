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
