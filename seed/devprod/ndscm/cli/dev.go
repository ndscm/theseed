package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdDev(args []string, ndConfig *common.NdConfig) error {
	if !seedshell.ShellEval() {
		log.Printf("\x1b[33mWarning: It's recommended to run nd-dev with --shell-eval\x1b[0m")
	}
	if ndConfig.Scm == "git" {
		err := common.QuickVerifyGitMonorepo(ndConfig)
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
		worktreePath := filepath.Join(ndConfig.MonorepoHome, worktreeName)
		worktreeStat, err := os.Stat(worktreePath)
		if err != nil && !os.IsNotExist(err) {
			return seederr.WrapErrorf("failed to stat worktree %v: %v", worktreePath, err)
		}
		if err == nil && !worktreeStat.IsDir() {
			return seederr.WrapErrorf("worktree %v exists and is not a dir", worktreePath)
		}
		if os.IsNotExist(err) {
			err = seedshell.ImpureRun("git", "--git-dir", ndConfig.MonorepoGitDir, "branch", "--track", "base/"+worktreeName, "origin/main")
			if err != nil {
				return seederr.WrapErrorf("failed to create base branch %v: %v", "base/"+worktreeName, err)
			}
			err = seedshell.ImpureRun("git", "--git-dir", ndConfig.MonorepoGitDir, "branch", "--track", worktreeName, "base/"+worktreeName)
			if err != nil {
				return seederr.WrapErrorf("failed to create worktree branch %v: %v", worktreeName, err)
			}
			err = seedshell.ImpureRun("git", "--git-dir", ndConfig.MonorepoGitDir, "worktree", "add", worktreePath, worktreeName)
			if err != nil {
				return seederr.WrapErrorf("failed to add worktree %v: %v", worktreePath, err)
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
	} else {
		return seederr.WrapErrorf("nd-dev does not support %v", ndConfig.Scm)
	}
	return nil
}
