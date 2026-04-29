package git

import (
	"strings"

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

func FetchAll(gitDir string) error {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "fetch", "--all", "--prune")...)
	if err != nil {
		return seederr.WrapErrorf("failed to fetch from all remotes: %w", err)
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

func ListRemoteBranches(gitDir string, remote string) ([]string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	pattern := "refs/remotes/" + remote + "/"
	output, err := seedshell.PureOutput("git", append(gitArgs, "for-each-ref", "--format=%(refname:short)", pattern)...)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to list remote branches for %v: %w", remote, err)
	}
	trimmed := strings.TrimSpace(string(output))
	if trimmed == "" {
		return nil, nil
	}
	lines := strings.Split(trimmed, "\n")
	branches := []string{}
	head := remote + "/HEAD"
	for _, line := range lines {
		if line != head {
			branches = append(branches, line)
		}
	}
	return branches, nil
}
