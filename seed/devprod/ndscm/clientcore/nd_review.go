package clientcore

import (
	"log"
	"os"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdReview(scmProvider scm.Provider, args []string) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-review should not run with --shell-eval")
	}
	if len(args) != 1 {
		return seederr.WrapErrorf("nd-review usage: nd review <feature-name>")
	}
	currentUserHandle := user.CurrentUserHandle()
	if currentUserHandle == "" {
		return seederr.WrapErrorf("user handle is not set")
	}
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.QuickVerifyMonorepo()
	if err != nil {
		return seederr.Wrap(err)
	}
	featureName := strings.TrimSpace(args[0])
	reviewBranch := "review/" + featureName
	changeBranch := "change/" + featureName
	worktreePath := scmProvider.GetBranchWorktree(monorepoHome, reviewBranch)
	_, err = os.Stat(worktreePath)
	if err != nil && !os.IsNotExist(err) {
		return seederr.Wrap(err)
	}
	if err == nil {
		log.Printf("Worktree %v already exists, removing...\n", worktreePath)
		err = scmProvider.RemoveWorktree(worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	_, err = scmProvider.GetCommitId(reviewBranch)
	if err == nil {
		log.Printf("Branch %v already exists, removing...\n", reviewBranch)
		err = scmProvider.DeleteBranch(reviewBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	trackingBranch, err := scmProvider.GetBranchTracking(changeBranch)
	if err != nil {
		return seederr.WrapErrorf("tracking upstream is missing for %v", changeBranch)
	}
	mergeBaseCommitId, err := scmProvider.GetMergeBaseCommitId("origin/main", changeBranch)
	if err != nil {
		return seederr.WrapErrorf("merge base is missing for %v", changeBranch)
	}
	err = scmProvider.CreateBranch(reviewBranch, mergeBaseCommitId, "origin/main")
	if err != nil {
		return seederr.Wrap(err)
	}
	worktreePath, err = scmProvider.CreateBranchWorktree(monorepoHome, reviewBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.ApplyCommitRange(worktreePath, trackingBranch, changeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.PushBranch(reviewBranch, "origin", currentUserHandle+"/"+featureName)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.RemoveWorktree(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(reviewBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
