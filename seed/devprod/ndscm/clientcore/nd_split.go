package clientcore

import (
	"errors"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdSplitOptions struct {
	Commit string
}

func NdSplit(scmProvider scm.Provider, options NdSplitOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-split should not run with --shell-eval")
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

	wipStatus, err := scmProvider.LoadWipStatus("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if wipStatus != nil {
		return seederr.WrapErrorf("wip status already exists")
	}

	wipBranchName, err := scm.GetWipBranchName(devWorktreeName)
	if err != nil {
		return seederr.Wrap(err)
	}
	_, err = scmProvider.GetBranch(wipBranchName)
	if err != nil && !errors.Is(err, scm.ErrBranchNotFound) {
		return seederr.Wrap(err)
	}
	if err == nil {
		return seederr.WrapErrorf("wip branch already exists")
	}

	targetCommitId, err := scmProvider.GetCommitId(options.Commit)
	if err != nil {
		return seederr.Wrap(err)
	}
	parentCommitId := ""
	currentBranch := devWorktreeName
	for !scm.IsBranchType(currentBranch, "base") {
		trackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if !scm.IsBranchType(trackingBranch, "change") &&
			!scm.IsBranchType(trackingBranch, "base") {
			return seederr.WrapErrorf("tracking chain is broken for %v (points to %v)", currentBranch, trackingBranch)
		}
		branchCommits, err := scmProvider.ListCommitIds(trackingBranch, currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		for i, commitId := range branchCommits {
			if targetCommitId != commitId {
				continue
			}
			if i > 0 {
				parentCommitId = branchCommits[i-1]
			} else {
				trackingCommitId, err := scmProvider.GetCommitId(trackingBranch)
				if err != nil {
					return seederr.Wrap(err)
				}
				parentCommitId = trackingCommitId
			}
			break
		}
		if parentCommitId != "" {
			break
		}
		currentBranch = trackingBranch
	}
	if parentCommitId == "" {
		return seederr.WrapErrorf("target %v (%v) does not exist on %v", options.Commit, targetCommitId, devWorktreeName)
	}

	err = scmProvider.CreateBranch(wipBranchName, parentCommitId, "")
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.Checkout("", wipBranchName)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.SaveWipStatus("", &scm.WipStatus{
		Operation: "split",
		Split: &scm.WipSplitStatus{
			Belong:   currentBranch,
			CommitId: targetCommitId,
		},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("Created wip at %v", parentCommitId)

	err = scmProvider.RestoreWorktree("", targetCommitId, false)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
