package clientcore

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdContinueOptions struct {
}

func continueSplit(scmProvider scm.Provider, wipStatus *scm.WipStatus, wipBranchName string) error {
	if wipStatus.Split == nil {
		return seederr.WrapErrorf("wip split status is missing a target commit")
	}

	// Commit whatever the user left uncommitted as the final split piece. If
	// the worktree is clean there is nothing left to commit, so skip it.
	dirtyFiles, err := scmProvider.ListDirtyFiles("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(dirtyFiles) > 0 {
		err = NdCommit(scmProvider, NdCommitOptions{Message: "split"})
		if err != nil {
			return seederr.Wrap(err)
		}
	}

	// Cap the split with a commit that reuses the original commit's message and
	// authorship, ensuring the wip tip matches the original commit exactly.
	err = scmProvider.RestoreWorktree("", wipStatus.Split.CommitId, true)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.CreateCommitReuse("", wipStatus.Split.CommitId)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Replay the commits that came after the original commit on its branch onto
	// the split, replacing the original commit with the split pieces.
	err = scmProvider.Checkout("", wipStatus.Split.Belong)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.Rebase("", wipStatus.Split.CommitId, wipBranchName)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(wipBranchName)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.RemoveWipStatus("", false)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = NdSync(scmProvider, NdSyncOptions{Fetch: "never"})
	if err != nil {
		return seederr.Wrap(err)
	}

	seedlog.Infof("Completed split of %v", wipStatus.Split.CommitId)
	return nil
}

func NdContinue(scmProvider scm.Provider, options NdContinueOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-continue should not run with --shell-eval")
	}

	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	worktreeName, _, err := scmProvider.GetCurrentWorktree(monorepoHome)
	if err != nil {
		return seederr.Wrap(err)
	}

	wipStatus, err := scmProvider.LoadWipStatus("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if wipStatus == nil {
		return seederr.WrapErrorf("no wip operation in progress")
	}

	wipBranchName, err := scm.GetWipBranchName(worktreeName)
	if err != nil {
		return seederr.Wrap(err)
	}
	currentBranch, err := scmProvider.GetWorktreeBranch("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if currentBranch != wipBranchName {
		return seederr.WrapErrorf("unexpected current branch. current=%v expected=%v", currentBranch, wipBranchName)
	}

	switch wipStatus.Operation {
	case "split":
		err := continueSplit(scmProvider, wipStatus, wipBranchName)
		if err != nil {
			return seederr.Wrap(err)
		}
	default:
		return seederr.WrapErrorf("unsupported wip operation: %v", wipStatus.Operation)
	}
	return nil
}
