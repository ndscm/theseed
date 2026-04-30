package git

import (
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

// BareClone creates a bare clone of repository at gitDir and returns the
// default branch name. A bare clone produces a .git directory without a
// default worktree, so we mirror the configs a normal clone would set
// (reflog, full remote fetch refspec, remote HEAD) to keep the gitDir
// usable as a backing store for dynamically-created worktrees.
func BareClone(repository string, gitDir string) (string, error) {
	err := seedshell.ImpureRun("git", "clone", "--bare", "--single-branch",
		"--config", "core.logallrefupdates=true",
		"--config", "remote.origin.fetch=+refs/heads/*:refs/remotes/origin/*",
		repository, gitDir)
	if err != nil {
		return "", seederr.WrapErrorf("failed to clone %v: %w", repository, err)
	}
	err = seedshell.ImpureRun("git", "--git-dir", gitDir, "remote", "set-head", "origin", "--auto")
	if err != nil {
		return "", seederr.WrapErrorf("failed to set remote HEAD for %v: %w", repository, err)
	}
	output, err := seedshell.PureOutput("git", "--git-dir", gitDir, "symbolic-ref", "refs/remotes/origin/HEAD")
	if err != nil {
		return "", seederr.WrapErrorf("failed to get default branch for remote origin: %w", err)
	}
	ref := strings.TrimSpace(string(output))
	prefix := "refs/remotes/origin/"
	defaultBranch := strings.TrimPrefix(ref, prefix)
	return defaultBranch, nil
}
