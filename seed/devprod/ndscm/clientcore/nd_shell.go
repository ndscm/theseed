package clientcore

import (
	"fmt"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore/shelleval"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdShell(args []string) error {
	seedflag.Finalize(args)
	if !seedshell.ShellEval() {
		seedlog.Warnf("It's recommended to run nd-shell with --shell-eval")
	}
	if len(args) != 0 {
		return seederr.WrapErrorf("nd-shell usage: eval \"$(ndscm --shell-eval shell)\"")
	}
	shellSnippets := strings.Join([]string{
		shelleval.NdSnippet(),
		shelleval.NdCompletionsSnippet(),
	}, "\n")
	if seedshell.Dry() {
		seedlog.Infof("Shell eval: %v", shellSnippets)
		return nil
	}
	if seedshell.ShellEval() {
		fmt.Printf("%v", shellSnippets)
	}
	return nil
}
