package main

import (
	"log"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
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
	if ndConfig.Scm == "git" {
		err := common.QuickVerifyGitMonorepo(ndConfig)
		if err != nil {
			return seederr.Wrap(err)
		}
		checkDirtyOutput, err := seedshell.PureOutput("git", "status", "--porcelain")
		if err != nil {
			return seederr.Wrap(err)
		}
		if strings.TrimSpace(string(checkDirtyOutput)) != "" {
			return seederr.WrapErrorf("workspace is dirty:\n%v", string(checkDirtyOutput))
		}
		devBranchOutput, err := seedshell.PureOutput("git", "branch", "--show-current")
		if err != nil {
			return seederr.Wrap(err)
		}
		devBranch := strings.TrimSpace(string(devBranchOutput))
		if !strings.HasPrefix(devBranch, "dev") {
			return seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranch)
		}
		worktreePathOutput, err := seedshell.PureOutput("git", "rev-parse", "--show-toplevel")
		if err != nil {
			return seederr.Wrap(err)
		}
		worktreePath := strings.TrimSpace(string(worktreePathOutput))
		worktree, err := filepath.Rel(ndConfig.MonorepoHome, worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		if devBranch != worktree {
			return seederr.WrapErrorf("worktree (%v) and dev branch (%v) mismatch", worktree, devBranch)
		}
		// # Iterate changes tree
		chain := []string{devBranch}
		for iter := devBranch; iter != ("base/" + devBranch); {
			inspectBranchOutput, err := seedshell.PureOutput("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", iter+"@{upstream}")
			if err != nil {
				return seederr.Wrap(err)
			}
			inspectBranch := strings.TrimSpace(string(inspectBranchOutput))
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
		err = seedshell.ImpureRun("git", "fetch", "--all", "--prune")
		if err != nil {
			return seederr.Wrap(err)
		}
		// # Rebase dev branch
		log.Printf("\x1b[34mRebasing: %v\x1b[0m", chain)
		incomingCommits := []string{}
		incomingCommitsOutput, err := seedshell.PureOutput("git", "rev-list", "base/"+devBranch+"..origin/main")
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, commitHash := range strings.Split(string(incomingCommitsOutput), "\n") {
			commitHash = strings.TrimSpace(commitHash)
			if commitHash != "" {
				incomingCommits = append(incomingCommits, commitHash)
			}
		}
		baseCommitHashOutput, err := seedshell.PureOutput("git", "rev-parse", "base/"+devBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		incomingCommits = append(incomingCommits, strings.TrimSpace(string(baseCommitHashOutput)))
		for i := len(incomingCommits) - 1; i >= 0; i-- {
			// Reverse iteration
			incommingCommitHash := incomingCommits[i]
			log.Printf("\x1b[34mRebasing to: %v\x1b[0m", incommingCommitHash)
			err := seedshell.ImpureRun("git", "checkout", "base/"+devBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = seedshell.ImpureRun("git", "rebase", incommingCommitHash)
			if err != nil {
				return seederr.Wrap(err)
			}
			for _, chainBranch := range chain[1:] {
				err := seedshell.ImpureRun("git", "checkout", chainBranch)
				if err != nil {
					return seederr.Wrap(err)
				}
				err = seedshell.ImpureRun("git", "pull", "--rebase")
				if err != nil {
					return seederr.Wrap(err)
				}
			}
			log.Printf("\x1b[32mRebased to: %v\x1b[0m", incommingCommitHash)
		}
		// # Cleanup local change branches
		for iter := devBranch; iter != ("base/" + devBranch); {
			inspectBranchOutput, err := seedshell.PureOutput("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", iter+"@{upstream}")
			if err != nil {
				return seederr.Wrap(err)
			}
			inspectBranch := strings.TrimSpace(string(inspectBranchOutput))
			if inspectBranch == "" {
				return seederr.WrapErrorf("tracking upstream is missing for %v", iter)
			}
			if !strings.HasPrefix(inspectBranch, "change/") && inspectBranch != ("base/"+devBranch) {
				return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", iter, inspectBranch)
			}
			if inspectBranch == ("base/" + devBranch) {
				break
			}
			nextBranchOutput, err := seedshell.PureOutput("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", inspectBranch+"@{upstream}")
			if err != nil {
				return seederr.Wrap(err)
			}
			nextBranch := strings.TrimSpace(string(nextBranchOutput))
			inspectCommitHashOutput, err := seedshell.PureOutput("git", "rev-parse", inspectBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			inspectCommitHash := strings.TrimSpace(string(inspectCommitHashOutput))
			nextCommitHashOutput, err := seedshell.PureOutput("git", "rev-parse", nextBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			nextCommitHash := strings.TrimSpace(string(nextCommitHashOutput))
			if inspectCommitHash == nextCommitHash {
				if !strings.HasPrefix(inspectBranch, "change/") {
					return seederr.WrapErrorf("unexpected empty branch %v", inspectBranch)
				}
				err := seedshell.ImpureRun("git", "branch", "-d", inspectBranch)
				if err != nil {
					return seederr.Wrap(err)
				}
				err = seedshell.ImpureRun("git", "branch", "--set-upstream-to="+nextBranch, iter)
				if err != nil {
					return seederr.Wrap(err)
				}
			} else {
				iter = inspectBranch
			}
		}
	} else {
		return seederr.WrapErrorf("nd-sync does not support %v", ndConfig.Scm)
	}
	return nil
}
