package clientcore

import (
	"fmt"
	"log"
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdDev(scmProvider scm.Provider, args []string) error {
	if !seedshell.ShellEval() {
		log.Printf("\x1b[33mWarning: It's recommended to run nd-dev with --shell-eval\x1b[0m")
	}
	scmName, err := scm.GetIdentifier(scmProvider)
	if err != nil {
		return seederr.Wrap(err)
	}
	switch scmName {
	case "git":
		monorepoHome, err := scm.MonorepoHome()
		if err != nil {
			return seederr.Wrap(err)
		}
		err = scmProvider.QuickVerifyMonorepo()
		if err != nil {
			return seederr.Wrap(err)
		}
		branchName := "dev"
		if len(args) == 0 {
			// pass
		} else if len(args) == 1 {
			branchName = "dev-" + args[0]
		} else {
			return seederr.WrapErrorf("nd-dev usage: nd dev [<feature-name>|<index>]")
		}
		worktreePath := scmProvider.GetBranchWorktree(monorepoHome, branchName)
		worktreeStat, err := os.Stat(worktreePath)
		if err != nil && !os.IsNotExist(err) {
			return seederr.WrapErrorf("failed to stat worktree %v: %v", worktreePath, err)
		}
		if err == nil && !worktreeStat.IsDir() {
			return seederr.WrapErrorf("worktree %v exists and is not a dir", worktreePath)
		}
		if os.IsNotExist(err) {
			newWorktreePath, err := scmProvider.CreateDevWorktree(monorepoHome, branchName)
			if err != nil {
				return seederr.WrapErrorf("failed to add branch worktree %v: %v", branchName, err)
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
