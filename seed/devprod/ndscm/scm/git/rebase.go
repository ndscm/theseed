package git

import (
	"os/exec"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

// ApplyFormatPatch applies a format-patch (as produced by GetFormatPatch and
// possibly rewritten) on top of the worktree's current branch with git am,
// feeding the patch in over stdin. --empty=keep preserves commits whose diffs
// were entirely stripped out, keeping their message and trailers intact.
func ApplyFormatPatch(worktreePath string, patch string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	gitArgs = append(gitArgs, "am", "--empty=keep")
	stdinOption := func(cmd *exec.Cmd) {
		cmd.Stdin = strings.NewReader(patch)
	}
	err := seedshell.ImpureOptionsRun([]seedshell.RunOption{stdinOption}, "git", gitArgs...)
	if err != nil {
		return seederr.WrapErrorf("failed to apply commit patch: %w", err)
	}
	return nil
}

func Rebase(worktreePath string, upstream string, onto string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	gitArgs = append(gitArgs, "rebase")
	if onto != "" {
		gitArgs = append(gitArgs, "--onto", onto)
	}
	gitArgs = append(gitArgs, upstream)
	err := seedshell.ImpureRun("git", gitArgs...)
	if err != nil {
		return seederr.WrapErrorf("failed to rebase (upstream %v) onto %v: %w", upstream, onto, err)
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
