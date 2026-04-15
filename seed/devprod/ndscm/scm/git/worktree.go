package git

import (
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func GetCurrentWorktreePath() (string, error) {
	worktreePathOutput, err := seedshell.PureOutput("git", "rev-parse", "--show-toplevel")
	if err != nil {
		return "", seederr.Wrap(err)
	}
	worktreePath := strings.TrimSpace(string(worktreePathOutput))
	return worktreePath, nil
}

func Checkout(worktreePath string, branchName string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "checkout", branchName)...)
	if err != nil {
		return seederr.WrapErrorf("failed to checkout branch %v: %w", branchName, err)
	}
	return nil
}

func CreateWorktree(gitDir string, worktreePath string, branchName string) error {
	if gitDir == "" {
		return seederr.WrapErrorf("git dir is required")
	}
	err := seedshell.ImpureRun("git", "--git-dir", gitDir, "worktree", "add", worktreePath, branchName)
	if err != nil {
		return seederr.WrapErrorf("failed to add worktree %v: %w", worktreePath, err)
	}
	return nil
}

func RemoveWorktree(gitDir string, worktreePath string) error {
	if gitDir == "" {
		return seederr.WrapErrorf("git dir is required")
	}
	err := seedshell.ImpureRun("git", "--git-dir", gitDir, "worktree", "remove", worktreePath)
	if err != nil {
		return seederr.WrapErrorf("failed to remove worktree %v: %w", worktreePath, err)
	}
	return nil
}

func CreateBranchWorktree(gitDir string, monorepoHome string, branchName string) (string, error) {
	if gitDir == "" {
		return "", seederr.WrapErrorf("git dir is required")
	}
	worktreePath := filepath.Join(monorepoHome, branchName)
	err := seedshell.ImpureRun("git", "--git-dir", gitDir, "worktree", "add", worktreePath, branchName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to add branch worktree %v: %w", worktreePath, err)
	}
	return worktreePath, nil
}

func GetBranchWorktreePath(monorepoHome string, branchName string) string {
	worktreePath := filepath.Join(monorepoHome, branchName)
	return worktreePath
}

func GetBranchWorktreeBranch(monorepoHome string, worktreePath string) (string, error) {
	branchName, err := filepath.Rel(monorepoHome, worktreePath)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if strings.HasPrefix(branchName, ".") {
		return "", seederr.WrapErrorf("worktree is not under monorepo home: %v", branchName)
	}
	return branchName, nil
}
