package git

import (
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func Rebase(worktreePath string, upstream string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "rebase", upstream)...)
	if err != nil {
		return seederr.WrapErrorf("failed to rebase onto %v: %w", upstream, err)
	}
	return nil
}

func PullRebase(worktreePath string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "pull", "--rebase")...)
	if err != nil {
		return seederr.WrapErrorf("failed to pull rebase: %w", err)
	}
	return nil
}

func CherryPickRange(worktreePath string, from string, to string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "cherry-pick", from+".."+to)...)
	if err != nil {
		return seederr.WrapErrorf("failed to cherry-pick range %v..%v: %w", from, to, err)
	}
	return nil
}

// SignOff rebases onto the configured tracking upstream, like PullRebase.
func SignOff(worktreePath string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "rebase", "--signoff")...)
	if err != nil {
		return seederr.WrapErrorf("failed to sign off branch: %w", err)
	}
	return nil
}
