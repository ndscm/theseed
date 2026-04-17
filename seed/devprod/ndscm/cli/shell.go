package main

import (
	"fmt"
	"log"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/cli/shelleval"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdShell(args []string) error {
	if !seedshell.ShellEval() {
		log.Printf("\x1b[33mWarning: It's recommended to run nd-shell with --shell-eval\x1b[0m")
	}
	if len(args) != 1 {
		return seederr.WrapErrorf("nd-shell usage: eval \"$(ndscm --shell-eval shell)\"")
	}
	shellSnippets := strings.Join([]string{
		shelleval.NdSnippet(),
		shelleval.NdCompletionsSnippet(),
	}, "\n")
	if seedshell.Dry() {
		log.Printf("Shell eval: %v", shellSnippets)
		return nil
	}
	if seedshell.ShellEval() {
		fmt.Printf("%v", shellSnippets)
	}
	return nil
}
