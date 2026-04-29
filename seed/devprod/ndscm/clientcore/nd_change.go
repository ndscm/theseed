package clientcore

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdChangeOptions struct {
	FeatureName string
}

func NdChange(scmProvider scm.Provider, options NdChangeOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-change should not run with --shell-eval")
	}
	devBranch, err := scmProvider.GetWorktreeBranch("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if !scmProvider.IsDevBranch(devBranch) {
		return seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranch)
	}
	changeBranch := "change/" + options.FeatureName
	currentBranch := devBranch
	for !strings.HasPrefix(currentBranch, "base/") {
		trackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if trackingBranch == changeBranch {
			// Navigation-only: dirty files intentionally carry over to the target branch.
			dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles("")
			if err != nil {
				return seederr.Wrap(err)
			}
			if len(dirtyFiles) > 0 {
				seedlog.Warnf("Carrying %v modification(s) to %v", len(dirtyFiles), changeBranch)
			}
			err = scmProvider.Checkout("", changeBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			return nil
		}
		if !strings.HasPrefix(trackingBranch, "change/") && !strings.HasPrefix(trackingBranch, "base/") {
			return seederr.WrapErrorf("tracking chain is broken for %v (points to %v)", currentBranch, trackingBranch)
		}
		currentBranch = trackingBranch
	}
	return seederr.WrapErrorf("change branch %v not found in tracking chain of %v", changeBranch, devBranch)
}
