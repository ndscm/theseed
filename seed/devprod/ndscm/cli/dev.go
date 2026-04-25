package main

import (
	"fmt"
	"log"
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm/git"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdDev(args []string) error {
	if !seedshell.ShellEval() {
		log.Printf("\x1b[33mWarning: It's recommended to run nd-dev with --shell-eval\x1b[0m")
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
		monorepoGitDir, err := git.MonorepoGitDir()
		if err != nil {
			return seederr.Wrap(err)
		}
		err = git.QuickVerifyMonorepo()
		if err != nil {
			return seederr.Wrap(err)
		}
		worktreeName := "dev"
		if len(args) == 1 {
			// pass
		} else if len(args) == 2 {
			worktreeName = "dev-" + args[1]
		} else {
			return seederr.WrapErrorf("nd-dev usage: nd dev [<feature-name>|<index>]")
		}
		worktreePath := git.GetBranchWorktreePath(monorepoHome, worktreeName)
		worktreeStat, err := os.Stat(worktreePath)
		if err != nil && !os.IsNotExist(err) {
			return seederr.WrapErrorf("failed to stat worktree %v: %v", worktreePath, err)
		}
		if err == nil && !worktreeStat.IsDir() {
			return seederr.WrapErrorf("worktree %v exists and is not a dir", worktreePath)
		}
		if os.IsNotExist(err) {
			err = git.CreateBranch(monorepoGitDir, "base/"+worktreeName, "origin/main", "origin/main")
			if err != nil {
				return seederr.WrapErrorf("failed to create base branch %v: %v", "base/"+worktreeName, err)
			}
			err = git.CreateBranch(monorepoGitDir, worktreeName, "base/"+worktreeName, "base/"+worktreeName)
			if err != nil {
				return seederr.WrapErrorf("failed to create worktree branch %v: %v", worktreeName, err)
			}
			newWorktreePath, err := git.CreateBranchWorktree(monorepoGitDir, monorepoHome, worktreeName)
			if err != nil {
				return seederr.WrapErrorf("failed to add branch worktree %v: %v", worktreeName, err)
			}
			if newWorktreePath != worktreePath {
				return seederr.WrapErrorf("unexpected new worktree path: %v (expected: %v)", newWorktreePath, worktreePath)
			}
		}
		shellEval := fmt.Sprintf("\ncd \"%v\"\n", worktreePath)
		if seedshell.Dry() {
			log.Printf("Shell eval: %v", shellEval)
		}
		if seedshell.ShellEval() {
			if !seedshell.Dry() {
				fmt.Printf("%v", shellEval)
			}
		}
	default:
		return seederr.WrapErrorf("nd-dev does not support %v", scmName)
	}
	return nil
}
