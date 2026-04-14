package main

import (
	"log"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdReview(args []string, ndConfig *common.NdConfig) error {
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
	if ndConfig.Scm == "git" {
		err := common.QuickVerifyGitMonorepo(ndConfig)
		if err != nil {
			return seederr.Wrap(err)
		}
		featureName := strings.TrimSpace(args[1])
		worktreePath := filepath.Join(ndConfig.MonorepoHome, "review/"+featureName)
		_, err = os.Stat(worktreePath)
		if err != nil && !os.IsNotExist(err) {
			return seederr.Wrap(err)
		}
		if err == nil {
			log.Printf("Worktree %v already exists, removing...\n", worktreePath)
			err = seedshell.ImpureRun("git", "worktree", "remove", worktreePath)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		_, err = seedshell.PureOutput("git", "rev-parse", "--verify", "review/"+featureName)
		if err == nil {
			log.Printf("Branch review/%v already exists, removing...\n", featureName)
			err = seedshell.ImpureRun("git", "branch", "--delete", "--force", "review/"+featureName)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
		trackingBranchOutput, err := seedshell.PureOutput("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", "change/"+featureName+"@{upstream}")
		if err != nil {
			return seederr.WrapErrorf("tracking upstream is missing for change/%v", featureName)
		}
		trackingBranch := strings.TrimSpace(string(trackingBranchOutput))
		mergeBaseHashOutput, err := seedshell.PureOutput("git", "merge-base", "origin/main", "change/"+featureName)
		if err != nil {
			return seederr.WrapErrorf("merge base is missing for change/%v", featureName)
		}
		mergeBaseHash := strings.TrimSpace(string(mergeBaseHashOutput))
		err = seedshell.ImpureRun("git", "branch", "review/"+featureName, mergeBaseHash)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = seedshell.ImpureRun("git", "branch", "--set-upstream-to=origin/main", "review/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = seedshell.ImpureRun("git", "worktree", "add", worktreePath, "review/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = seedshell.ImpureRun("git", "-C", worktreePath, "cherry-pick", trackingBranch+"..change/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = seedshell.ImpureRun("git", "push", "--force", "origin", "review/"+featureName+":"+currentUserHandle+"/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = seedshell.ImpureRun("git", "worktree", "remove", worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = seedshell.ImpureRun("git", "branch", "--delete", "--force", "review/"+featureName)
		if err != nil {
			return seederr.Wrap(err)
		}
	} else {
		return seederr.WrapErrorf("nd-review does not support %v", ndConfig.Scm)
	}
	return nil
}
