package main

import (
	"fmt"
	"log"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
)

func NdSync(args []string, ndConfig *common.NdConfig) error {
	if ndConfig.ShellEval {
		return common.WrapTrace(fmt.Errorf("nd-sync should not run with --shell-eval"))
	}
	if len(args) != 1 {
		return common.WrapTrace(fmt.Errorf("nd-sync usage: (on dev branch) nd sync"))
	}
	if ndConfig.Scm == "git" {
		err := common.QuickVerifyGitMonorepo(ndConfig)
		if err != nil {
			return err
		}
		checkDirtyOutput, err := common.ShellOutput(false, "git", "status", "--porcelain")
		if err != nil {
			return err
		}
		if strings.TrimSpace(string(checkDirtyOutput)) != "" {
			return common.WrapTrace(fmt.Errorf("workspace is dirty:\n%v", string(checkDirtyOutput)))
		}
		devBranchOutput, err := common.ShellOutput(false, "git", "branch", "--show-current")
		if err != nil {
			return err
		}
		devBranch := strings.TrimSpace(string(devBranchOutput))
		if !strings.HasPrefix(devBranch, "dev") {
			return common.WrapTrace(fmt.Errorf("workspace branch is not a dev branch: %v", devBranch))
		}
		worktreePathOutput, err := common.ShellOutput(false, "git", "rev-parse", "--show-toplevel")
		if err != nil {
			return err
		}
		worktreePath := strings.TrimSpace(string(worktreePathOutput))
		worktree, err := filepath.Rel(ndConfig.MonorepoHome, worktreePath)
		if err != nil {
			return common.WrapTrace(err)
		}
		if devBranch != worktree {
			return common.WrapTrace(fmt.Errorf("worktree (%v) and dev branch (%v) mismatch", worktree, devBranch))
		}
		// # Iterate changes tree
		chain := []string{devBranch}
		for iter := devBranch; iter != ("base/" + devBranch); {
			inspectBranchOutput, err := common.ShellOutput(false, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", iter+"@{upstream}")
			if err != nil {
				return err
			}
			inspectBranch := strings.TrimSpace(string(inspectBranchOutput))
			if inspectBranch == "" {
				return common.WrapTrace(fmt.Errorf("tracking upstream is missing for %v", iter))
			}
			if !strings.HasPrefix(inspectBranch, "change/") && inspectBranch != ("base/"+devBranch) {
				return common.WrapTrace(fmt.Errorf("tracking chain is broken for %v (point to %v)", iter, inspectBranch))
			}
			chain = append([]string{inspectBranch}, chain...)
			iter = inspectBranch
		}
		// # Fetch upstream changes
		err = common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "fetch", "--all", "--prune")
		if err != nil {
			return err
		}
		// # Rebase dev branch
		log.Printf("\x1b[34mRebasing: %v\x1b[0m", chain)
		incomingCommitsOutput, err := common.ShellOutput(false, "git", "rev-list", "base/"+devBranch+"..origin/main")
		if err != nil {
			return err
		}
		baseCommitHashOutput, err := common.ShellOutput(false, "git", "rev-parse", "base/"+devBranch)
		if err != nil {
			return err
		}
		incomingCommits := strings.Split(
			strings.TrimSpace(string(incomingCommitsOutput))+
				"\n"+
				strings.TrimSpace(string(baseCommitHashOutput)),
			"\n")
		for i, commitHash := range incomingCommits {
			incomingCommits[i] = strings.TrimSpace(commitHash)
		}
		for i := len(incomingCommits) - 1; i >= 0; i-- {
			// Reverse iteration
			incommingCommitHash := incomingCommits[i]
			log.Printf("\x1b[34mRebasing to: %v\x1b[0m", incommingCommitHash)
			err := common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "checkout", "base/"+devBranch)
			if err != nil {
				return err
			}
			err = common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "rebase", incommingCommitHash)
			if err != nil {
				return err
			}
			for _, chainBranch := range chain[1:] {
				err := common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "checkout", chainBranch)
				if err != nil {
					return err
				}
				err = common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "pull", "--rebase")
				if err != nil {
					return err
				}
			}
			log.Printf("\x1b[32mRebased to: %v\x1b[0m", incommingCommitHash)
		}
		// # Cleanup local change branches
		for iter := devBranch; iter != ("base/" + devBranch); {
			inspectBranchOutput, err := common.ShellOutput(false, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", iter+"@{upstream}")
			if err != nil {
				return err
			}
			inspectBranch := strings.TrimSpace(string(inspectBranchOutput))
			if inspectBranch == "" {
				return common.WrapTrace(fmt.Errorf("tracking upstream is missing for %v", iter))
			}
			if !strings.HasPrefix(inspectBranch, "change/") && inspectBranch != ("base/"+devBranch) {
				return common.WrapTrace(fmt.Errorf("tracking chain is broken for %v (point to %v)", iter, inspectBranch))
			}
			if inspectBranch == ("base/" + devBranch) {
				break
			}
			nextBranchOutput, err := common.ShellOutput(false, "git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", inspectBranch+"@{upstream}")
			if err != nil {
				return err
			}
			nextBranch := strings.TrimSpace(string(nextBranchOutput))
			inspectCommitHashOutput, err := common.ShellOutput(false, "git", "rev-parse", inspectBranch)
			if err != nil {
				return err
			}
			inspectCommitHash := strings.TrimSpace(string(inspectCommitHashOutput))
			nextCommitHashOutput, err := common.ShellOutput(false, "git", "rev-parse", nextBranch)
			if err != nil {
				return err
			}
			nextCommitHash := strings.TrimSpace(string(nextCommitHashOutput))
			if inspectCommitHash == nextCommitHash {
				if !strings.HasPrefix(inspectBranch, "change/") {
					return common.WrapTrace(fmt.Errorf("unexpected empty branch %v", inspectBranch))
				}
				err := common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "branch", "-d", inspectBranch)
				if err != nil {
					return err
				}
				err = common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "branch", "--set-upstream-to="+nextBranch, iter)
				if err != nil {
					return err
				}
			} else {
				iter = inspectBranch
			}
		}
	} else {
		return common.WrapTrace(fmt.Errorf("nd-sync does not support %v", ndConfig.Scm))
	}
	return nil
}
