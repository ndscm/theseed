package clientcore

import (
	"errors"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdCutOptions struct {
	Force bool

	FeatureName string
	CutPoint    string
}

func NdCut(scmProvider scm.Provider, options NdCutOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-cut should not run with --shell-eval")
	}

	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	dirtyFiles, err := scmProvider.ListDirtyFiles("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(dirtyFiles) > 0 {
		return seederr.WrapErrorf("workspace is dirty:\n%v", dirtyFiles)
	}
	devWorktreeName, _, err := scmProvider.GetCurrentWorktree(monorepoHome)
	if err != nil {
		return seederr.Wrap(err)
	}
	if !scm.IsBranchType(devWorktreeName, "dev") {
		return seederr.WrapErrorf("current worktree is not a dev worktree: %v", devWorktreeName)
	}

	if strings.Contains(options.FeatureName, "/") {
		return seederr.WrapErrorf("feature name %v must not contain /", options.FeatureName)
	}

	changeBranch, err := scm.ChangeBranchName(devWorktreeName, options.FeatureName)
	if err != nil {
		return seederr.Wrap(err)
	}
	_, err = scmProvider.GetBranch(changeBranch)
	if err != nil && !errors.Is(err, scm.ErrBranchNotFound) {
		return seederr.Wrap(err)
	}
	if err == nil {
		if !options.Force {
			return seederr.WrapErrorf("change branch %v already exists, use --force to replace", changeBranch)
		}
		seedlog.Warnf("Branch %v already exists, removing...", changeBranch)
		err = NdUncut(scmProvider, NdUncutOptions{FeatureName: options.FeatureName})
		if err != nil {
			return seederr.Wrap(err)
		}
	}

	targetCommitId, err := scmProvider.GetCommitId(options.CutPoint)
	if err != nil {
		return seederr.Wrap(err)
	}
	currentBranch := devWorktreeName
	trackingBranch := ""
	for !scm.IsBranchType(currentBranch, "base") {
		currentTrackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if !scm.IsBranchType(currentTrackingBranch, "change") &&
			!scm.IsBranchType(currentTrackingBranch, "base") {
			return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", currentBranch, currentTrackingBranch)
		}
		branchCommits, err := scmProvider.ListCommitIds(currentTrackingBranch, currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, commitId := range branchCommits {
			if targetCommitId == commitId {
				trackingBranch = currentTrackingBranch
				break
			}
		}
		if trackingBranch != "" {
			break
		}
		currentBranch = currentTrackingBranch
	}
	if trackingBranch == "" {
		return seederr.WrapErrorf("target %v (%v) does not exist on %v", options.CutPoint, targetCommitId, devWorktreeName)
	}
	err = scmProvider.CreateBranch(changeBranch, targetCommitId, trackingBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.SetBranchTracking(currentBranch, changeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("Created change request as %v", changeBranch)
	return nil
}
