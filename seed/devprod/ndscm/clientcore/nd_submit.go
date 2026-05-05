package clientcore

import (
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdSubmitOptions struct {
	Force  bool
	Remote string

	FeatureName string
	CutPoint    string
}

func NdSubmit(scmProvider scm.Provider, options NdSubmitOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-submit should not run with --shell-eval")
	}
	remoteMain := options.Remote + "/main"
	currentUserHandle, err := user.CurrentUserHandle()
	if err != nil {
		return seederr.Wrap(err)
	}
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.QuickVerifyMonorepo()
	if err != nil {
		return seederr.Wrap(err)
	}
	if options.CutPoint != "" {
		err = NdCut(scmProvider, NdCutOptions{
			Force:       options.Force,
			FeatureName: options.FeatureName,
			CutPoint:    options.CutPoint,
		})
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	submitBranch := "submit/" + options.FeatureName
	changeBranch := "change/" + options.FeatureName
	worktreePath := scmProvider.GetBranchWorktree(monorepoHome, submitBranch)
	_, err = os.Stat(worktreePath)
	if err != nil && !os.IsNotExist(err) {
		return seederr.Wrap(err)
	}
	if err == nil {
		seedlog.Warnf("Worktree %v already exists, removing...", worktreePath)
		err = scmProvider.RemoveWorktree(worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	_, err = scmProvider.GetCommitId(submitBranch)
	if err == nil {
		seedlog.Warnf("Branch %v already exists, removing...", submitBranch)
		err = scmProvider.DeleteBranch(submitBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	trackingBranch, err := scmProvider.GetBranchTracking(changeBranch)
	if err != nil {
		return seederr.WrapErrorf("tracking upstream is missing for %v", changeBranch)
	}
	mergeBaseCommitId, err := scmProvider.GetMergeBaseCommitId(remoteMain, changeBranch)
	if err != nil {
		return seederr.WrapErrorf("merge base is missing for %v", changeBranch)
	}
	err = scmProvider.CreateBranch(submitBranch, mergeBaseCommitId, remoteMain)
	if err != nil {
		return seederr.Wrap(err)
	}
	worktreePath, err = scmProvider.CreateBranchWorktree(monorepoHome, submitBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.ApplyCommitRange(worktreePath, trackingBranch, changeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.SignOff(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.PushBranch(submitBranch, options.Remote, currentUserHandle+"/"+options.FeatureName)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.RemoveWorktree(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(submitBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
