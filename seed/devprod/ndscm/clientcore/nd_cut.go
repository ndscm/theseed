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
	currentBranch := devBranch
	for !strings.HasPrefix(currentBranch, "base/") {
		trackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if !strings.HasPrefix(trackingBranch, "change/") && !strings.HasPrefix(trackingBranch, "base/") {
			return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", currentBranch, trackingBranch)
		}
		branchCommits, err := scmProvider.ListCommitIds(trackingBranch, currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		found := false
		for _, commitId := range branchCommits {
			if targetCommitId == commitId {
				found = true
				break
			}
		}
		if found {
			err := scmProvider.CreateBranch(changeBranch, targetCommitId, trackingBranch)
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
		currentBranch = trackingBranch
	}
	return seederr.WrapErrorf("target %v (%v) does not exist on %v branch", options.CutPoint, targetCommitId, devBranch)
}
