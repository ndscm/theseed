package main

import (
	"fmt"
	"log"
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
)

func NdDev(args []string, ndConfig *common.NdConfig) error {
	if !ndConfig.ShellEval {
		log.Printf("\x1b[33mWarning: It's recommended to run nd-dev with --shell-eval\x1b[0m")
	}
	if ndConfig.Scm == "git" {
		err := common.QuickVerifyGitMonorepo(ndConfig)
		if err != nil {
			return err
		}
		worktreeName := "dev"
		if len(args) == 1 {
			// pass
		} else if len(args) == 2 {
			worktreeName = "dev-" + args[1]
		} else {
			return common.WrapTrace(fmt.Errorf("nd-dev usage: nd dev [<feature-name>|<index>]"))
		}
		worktreePath := filepath.Join(ndConfig.MonorepoHome, worktreeName)
		worktreeStat, err := os.Stat(worktreePath)
		if err != nil && !os.IsNotExist(err) {
			return common.WrapTrace(err)
		}
		if err == nil && !worktreeStat.IsDir() {
			return common.WrapTrace(fmt.Errorf("worktree %v exists and is not a dir", worktreePath))
		}
		if os.IsNotExist(err) {
			err = common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "--git-dir", ndConfig.MonorepoGitDir, "branch", "--track=direct", "base/"+worktreeName, "origin/main")
			if err != nil {
				return err
			}
			err = common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "--git-dir", ndConfig.MonorepoGitDir, "branch", "--track=direct", worktreeName, "base/"+worktreeName)
			if err != nil {
				return err
			}
			err = common.ShellRun(ndConfig.Dry, ndConfig.ShellEval, "git", "--git-dir", ndConfig.MonorepoGitDir, "worktree", "add", worktreePath, worktreeName)
			if err != nil {
				return err
			}
		}
		shellEval := fmt.Sprintf("\ncd \"%v\"\n", worktreePath)
		if ndConfig.Dry {
			log.Printf("Shell eval: %v", shellEval)
		}
		if ndConfig.ShellEval {
			if !ndConfig.Dry {
				fmt.Printf("%v", shellEval)
			}
		}
	} else {
		return common.WrapTrace(fmt.Errorf("nd-dev does not support %v", ndConfig.Scm))
	}
	return nil
}
