package clientcore

import (
	"slices"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdSyncOptions struct {
}

func NdSync(scmProvider scm.Provider, _ NdSyncOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-sync should not run with --shell-eval")
	}

	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.QuickVerifyMonorepo()
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
	worktreeName, _, err := scmProvider.GetCurrentWorktree(monorepoHome)
	if err != nil {
		return seederr.Wrap(err)
	}
	if !scm.IsBranchType(worktreeName, "dev") &&
		!scm.IsBranchType(worktreeName, "melt") {
		return seederr.WrapErrorf("workspace is not a ndscm managed worktree: %v", worktreeName)
	}

	baseBranch, err := scm.GetBaseBranchName(worktreeName)
	if err != nil {
		return seederr.Wrap(err)
	}
	// # Iterate changes tree
	chain := []string{worktreeName}
	for iter := worktreeName; iter != baseBranch; {
		inspectBranch, err := scmProvider.GetBranchTracking(iter)
		if err != nil {
			return seederr.Wrap(err)
		}
		if inspectBranch == "" {
			return seederr.WrapErrorf("tracking upstream is missing for %v", iter)
		}
		if !scm.IsBranchType(inspectBranch, "change") && inspectBranch != baseBranch {
			return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", iter, inspectBranch)
		}
		chain = append([]string{inspectBranch}, chain...)
		iter = inspectBranch
	}
	activeBranch, err := scmProvider.GetWorktreeBranch("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if !slices.Contains(chain, activeBranch) {
		return seederr.WrapErrorf("active branch %v is not in the changes chain: %v", activeBranch, chain)
	}
	// # Fetch upstream changes
	err = scmProvider.FetchAll()
	if err != nil {
		return seederr.Wrap(err)
	}
	// # Rebase dev branch
	baseTracking, err := scmProvider.GetBranchTracking(baseBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("\x1b[34mRebasing onto %v: %v\x1b[0m", baseTracking, chain)
	mergeBaseCommitId, err := scmProvider.GetMergeBaseCommitId(baseBranch, baseTracking)
	if err != nil {
		return seederr.Wrap(err)
	}
	incomingCommits := []string{mergeBaseCommitId}
	trackingIncomingCommits, err := scmProvider.ListCommitIds(mergeBaseCommitId, baseTracking)
	if err != nil {
		return seederr.Wrap(err)
	}
	incomingCommits = append(incomingCommits, trackingIncomingCommits...)
	for _, incomingCommitId := range incomingCommits {
		seedlog.Infof("\x1b[34mRebasing to: %v\x1b[0m", incomingCommitId)
		err := scmProvider.Checkout("", baseBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = scmProvider.Rebase("", incomingCommitId, "")
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, chainBranch := range chain[1:] {
			err := scmProvider.Checkout("", chainBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = scmProvider.PullRebase("")
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		seedlog.Infof("\x1b[32mRebased to: %v\x1b[0m", incomingCommitId)
	}
	// # Cleanup local change branches
	for iter := worktreeName; iter != baseBranch; {
		inspectBranch, err := scmProvider.GetBranchTracking(iter)
		if err != nil {
			return seederr.Wrap(err)
		}
		if inspectBranch == "" {
			return seederr.WrapErrorf("tracking upstream is missing for %v", iter)
		}
		if !scm.IsBranchType(inspectBranch, "change") && inspectBranch != baseBranch {
			return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", iter, inspectBranch)
		}
		if inspectBranch == baseBranch {
			break
		}
		nextBranch, err := scmProvider.GetBranchTracking(inspectBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		inspectCommitId, err := scmProvider.GetCommitId(inspectBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		nextCommitId, err := scmProvider.GetCommitId(nextBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if inspectCommitId == nextCommitId {
			if !scm.IsBranchType(inspectBranch, "change") {
				return seederr.WrapErrorf("unexpected empty branch %v", inspectBranch)
			}
			err := scmProvider.DeleteMergedBranch(inspectBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = scmProvider.SetBranchTracking(iter, nextBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
		} else {
			iter = inspectBranch
		}
	}
	return nil
}
