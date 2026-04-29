package clientcore

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdUncutOptions struct {
	FeatureName string
}

func NdUncut(scmProvider scm.Provider, options NdUncutOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-uncut should not run with --shell-eval")
	}
	dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(dirtyFiles) > 0 {
		return seederr.WrapErrorf("workspace is dirty:\n%v", dirtyFiles)
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
	childBranch := ""
	for !strings.HasPrefix(currentBranch, "base/") {
		currentTrackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if currentTrackingBranch == changeBranch {
			childBranch = currentBranch
			break
		}
		if !strings.HasPrefix(currentTrackingBranch, "change/") && !strings.HasPrefix(currentTrackingBranch, "base/") {
			return seederr.WrapErrorf("tracking chain is broken for %v (points to %v)", currentBranch, currentTrackingBranch)
		}
		currentBranch = currentTrackingBranch
	}
	if childBranch == "" {
		return seederr.WrapErrorf("change branch %v not found in tracking chain of %v", changeBranch, devBranch)
	}
	changeTracking, err := scmProvider.GetBranchTracking(changeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.SetBranchTracking(childBranch, changeTracking)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(changeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("Removed change request %v", changeBranch)
	return nil
}
