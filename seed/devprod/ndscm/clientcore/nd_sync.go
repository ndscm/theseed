package clientcore

import (
	"slices"
	"strings"

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
	dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(dirtyFiles) > 0 {
		return seederr.WrapErrorf("workspace is dirty:\n%v", dirtyFiles)
	}
	worktreePath, err := scmProvider.GetCurrentWorktree()
	if err != nil {
		return seederr.Wrap(err)
	}
	devBranch, err := scmProvider.GetBranchWorktreeBranch(monorepoHome, worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	if !scmProvider.IsDevBranch(devBranch) {
		return seederr.WrapErrorf("workspace is not a dev worktree: %v", devBranch)
	}
	// # Iterate changes tree
	chain := []string{devBranch}
	for iter := devBranch; iter != ("base/" + devBranch); {
		inspectBranch, err := scmProvider.GetBranchTracking(iter)
		if err != nil {
			return seederr.Wrap(err)
		}
		if inspectBranch == "" {
			return seederr.WrapErrorf("tracking upstream is missing for %v", iter)
		}
		if !strings.HasPrefix(inspectBranch, "change/") && inspectBranch != ("base/"+devBranch) {
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
	baseTracking, err := scmProvider.GetBranchTracking("base/" + devBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("\x1b[34mRebasing onto %v: %v\x1b[0m", baseTracking, chain)
	incomingCommits, err := scmProvider.ListCommitIds("base/"+devBranch, baseTracking)
	if err != nil {
		return seederr.Wrap(err)
	}
	baseCommitId, err := scmProvider.GetCommitId("base/" + devBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	incomingCommits = append(incomingCommits, baseCommitId)
	for i := len(incomingCommits) - 1; i >= 0; i-- {
		// Reverse iteration
		incomingCommitId := incomingCommits[i]
		seedlog.Infof("\x1b[34mRebasing to: %v\x1b[0m", incomingCommitId)
		err := scmProvider.Checkout("", "base/"+devBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = scmProvider.Rebase("", incomingCommitId)
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
	for iter := devBranch; iter != ("base/" + devBranch); {
		inspectBranch, err := scmProvider.GetBranchTracking(iter)
		if err != nil {
			return seederr.Wrap(err)
		}
		if inspectBranch == "" {
			return seederr.WrapErrorf("tracking upstream is missing for %v", iter)
		}
		if !strings.HasPrefix(inspectBranch, "change/") && inspectBranch != ("base/"+devBranch) {
			return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", iter, inspectBranch)
		}
		if inspectBranch == ("base/" + devBranch) {
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
			if !strings.HasPrefix(inspectBranch, "change/") {
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
