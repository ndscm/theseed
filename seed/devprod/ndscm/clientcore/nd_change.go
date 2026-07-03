package clientcore

import (
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

	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	devWorktreeName, _, err := scmProvider.GetCurrentWorktree(monorepoHome)
	if err != nil {
		return seederr.Wrap(err)
	}
	if !scm.IsBranchType(devWorktreeName, "dev") {
		return seederr.WrapErrorf("current worktree is not a dev worktree: %v", devWorktreeName)
	}

	changeBranch, err := scm.GetChangeBranchName(devWorktreeName, options.FeatureName)
	if err != nil {
		return seederr.Wrap(err)
	}
	currentBranch := devWorktreeName
	for !scm.IsBranchType(currentBranch, "base") {
		trackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if trackingBranch == changeBranch {
			// Navigation-only: dirty files intentionally carry over to the target branch.
			dirtyFiles, err := scmProvider.ListDirtyFiles("")
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
		if !scm.IsBranchType(trackingBranch, "change") &&
			!scm.IsBranchType(trackingBranch, "base") {
			return seederr.WrapErrorf("tracking chain is broken for %v (points to %v)", currentBranch, trackingBranch)
		}
		currentBranch = trackingBranch
	}
	return seederr.WrapErrorf("change branch %v not found in tracking chain of %v", changeBranch, devWorktreeName)
}
