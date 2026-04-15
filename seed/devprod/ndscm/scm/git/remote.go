package git

import (
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func Fetch(gitDir string, remote string) error {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "fetch", remote, "--prune")...)
	if err != nil {
		return seederr.WrapErrorf("failed to fetch from remote %v: %w", remote, err)
	}
	return nil
}

func PushBranch(gitDir string, branchName string, remote string, remoteBranchName string) error {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "push", "--force", remote, branchName+":"+remoteBranchName)...)
	if err != nil {
		return seederr.WrapErrorf("failed to push branch %v to remote %v: %w", branchName, remote, err)
	}
	return nil
}
