package main

import (
	"log"
	"os"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm/git"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdReview(args []string) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-review should not run with --shell-eval")
	}
	if len(args) != 2 {
		return seederr.WrapErrorf("nd-review usage: nd review <feature-name>")
	}
	currentUserHandle := user.CurrentUserHandle()
	if currentUserHandle == "" {
		return seederr.WrapErrorf("user handle is not set")
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
		monorepoGitDir, err := scm.MonorepoGitDir()
		if err != nil {
			return seederr.Wrap(err)
		}
		err = common.QuickVerifyGitMonorepo()
		if err != nil {
			return seederr.Wrap(err)
		}
		featureName := strings.TrimSpace(args[1])
		worktreePath := git.GetBranchWorktreePath(monorepoHome, "review/"+featureName)
		_, err = os.Stat(worktreePath)
		if err != nil && !os.IsNotExist(err) {
			return seederr.Wrap(err)
		}
		if err == nil {
			log.Printf("Worktree %v already exists, removing...\n", worktreePath)
			err = git.RemoveWorktree(monorepoGitDir, worktreePath)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		_, err = git.GetCommitHash(monorepoGitDir, "review/"+featureName)
		if err == nil {
			log.Printf("Branch review/%v already exists, removing...\n", featureName)
			err = git.DeleteBranch(monorepoGitDir, "review/"+featureName)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		trackingBranch, err := git.GetBranchTracking(monorepoGitDir, "change/"+featureName)
		if err != nil {
			return seederr.WrapErrorf("tracking upstream is missing for change/%v", featureName)
		}
		mergeBaseHash, err := git.GetMergeBaseHash(monorepoGitDir, "origin/main", "change/"+featureName)
		if err != nil {
			return seederr.WrapErrorf("merge base is missing for change/%v", featureName)
		}
		err = git.CreateBranch(monorepoGitDir, "review/"+featureName, mergeBaseHash, "origin/main")
		if err != nil {
			return seederr.Wrap(err)
		}
		worktreePath, err = git.CreateBranchWorktree(monorepoGitDir, monorepoHome, "review/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = git.CherryPickRange(worktreePath, trackingBranch, "change/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = git.PushBranch(monorepoGitDir, "review/"+featureName, "origin", currentUserHandle+"/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = git.RemoveWorktree(monorepoGitDir, worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = git.DeleteBranch(monorepoGitDir, "review/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
	default:
		return seederr.WrapErrorf("nd-review does not support %v", scmName)
	}
	return nil
}
