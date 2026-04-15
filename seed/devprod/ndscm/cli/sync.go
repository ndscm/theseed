package main

import (
	"log"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm/git"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdSync(args []string, ndConfig *common.NdConfig) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-sync should not run with --shell-eval")
	}
	if len(args) != 1 {
		return seederr.WrapErrorf("nd-sync usage: (on dev branch) nd sync")
	}
	scmName, err := scm.ScmName()
	if err != nil {
		return seederr.Wrap(err)
	}
	switch scmName {
	case "git":
		monorepoHome, err := scm.MonorepoHome()
		if err != nil {
			return seederr.Wrap(err)
		}
		err = common.QuickVerifyGitMonorepo()
		if err != nil {
			return seederr.Wrap(err)
		}
		dirtyFiles, err := git.GetStatus("")
		if err != nil {
			return seederr.Wrap(err)
		}
		if len(dirtyFiles) > 0 {
			return seederr.WrapErrorf("workspace is dirty:\n%v", dirtyFiles)
		}
		devBranch, err := git.GetCurrentBranch("")
		if err != nil {
			return seederr.Wrap(err)
		}
		if !strings.HasPrefix(devBranch, "dev") {
			return seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranch)
		}
		worktreePath, err := git.GetCurrentWorktreePath()
		if err != nil {
			return seederr.Wrap(err)
		}
		worktree, err := git.GetBranchWorktreeBranch(monorepoHome, worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		if devBranch != worktree {
			return seederr.WrapErrorf("worktree (%v) and dev branch (%v) mismatch", worktree, devBranch)
		}
		// # Iterate changes tree
		chain := []string{devBranch}
		for iter := devBranch; iter != ("base/" + devBranch); {
			inspectBranch, err := git.GetBranchTracking("", iter)
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
		// # Fetch upstream changes
		err = git.FetchAll("")
		if err != nil {
			return seederr.Wrap(err)
		}
		// # Rebase dev branch
		log.Printf("\x1b[34mRebasing: %v\x1b[0m", chain)
		incomingCommits, err := git.ListCommitHash("", "base/"+devBranch, "origin/main")
		if err != nil {
			return seederr.Wrap(err)
		}
		baseCommitHash, err := git.GetCommitHash("", "base/"+devBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		incomingCommits = append(incomingCommits, baseCommitHash)
		for i := len(incomingCommits) - 1; i >= 0; i-- {
			// Reverse iteration
			incommingCommitHash := incomingCommits[i]
			log.Printf("\x1b[34mRebasing to: %v\x1b[0m", incommingCommitHash)
			err := git.Checkout("", "base/"+devBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = git.Rebase("", incommingCommitHash)
			if err != nil {
				return seederr.Wrap(err)
			}
			for _, chainBranch := range chain[1:] {
				err := git.Checkout("", chainBranch)
				if err != nil {
					return seederr.Wrap(err)
				}
				err = git.PullRebase("")
				if err != nil {
					return seederr.Wrap(err)
				}
			}
			log.Printf("\x1b[32mRebased to: %v\x1b[0m", incommingCommitHash)
		}
		// # Cleanup local change branches
		for iter := devBranch; iter != ("base/" + devBranch); {
			inspectBranch, err := git.GetBranchTracking("", iter)
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
			nextBranch, err := git.GetBranchTracking("", inspectBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			inspectCommitHash, err := git.GetCommitHash("", inspectBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			nextCommitHash, err := git.GetCommitHash("", nextBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			if inspectCommitHash == nextCommitHash {
				if !strings.HasPrefix(inspectBranch, "change/") {
					return seederr.WrapErrorf("unexpected empty branch %v", inspectBranch)
				}
				err := git.DeleteMergedBranch("", inspectBranch)
				if err != nil {
					return seederr.Wrap(err)
				}
				err = git.SetBranchTracking("", iter, nextBranch)
				if err != nil {
					return seederr.Wrap(err)
				}
			} else {
				iter = inspectBranch
			}
		}
	default:
		return seederr.WrapErrorf("nd-sync does not support %v", scmName)
	}
	return nil
}
