package clientcore

import (
	"os"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type ndSubmitFlags struct {
	remote *seedflag.StringFlag
}

func parseNdSubmitFlags(args []string) (ndSubmitFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-submit")
	cmdFlags := ndSubmitFlags{}
	cmdFlags.remote = cf.DefineString("remote", "origin", "Remote identifier for submitting the branch")
	cmdArgs, err := cf.Parse(args)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func NdSubmit(scmProvider scm.Provider, args []string) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-submit should not run with --shell-eval")
	}
	cmdFlags, cmdArgs, err := parseNdSubmitFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	if len(cmdArgs) != 1 {
		return seederr.WrapErrorf("nd-submit usage: nd submit [...flags] <feature-name>")
	}
	featureName := strings.TrimSpace(cmdArgs[0])
	remoteMain := cmdFlags.remote.Get() + "/main"
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
	submitBranch := "submit/" + featureName
	changeBranch := "change/" + featureName
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
	err = scmProvider.PushBranch(submitBranch, cmdFlags.remote.Get(), currentUserHandle+"/"+featureName)
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
