package clientcore

import (
	"fmt"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore/shelleval"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdShellOptions struct {
}

func NdShell(_ NdShellOptions) error {
	shellSnippets := strings.Join([]string{
		shelleval.NdSnippet(),
		shelleval.NdCompletionsSnippet(),
	}, "\n")
	if seedshell.Dry() {
		seedlog.Infof("Shell eval: %v", shellSnippets)
		return nil
	} else {
		fmt.Printf("%v", shellSnippets)
	}
	return nil
}
