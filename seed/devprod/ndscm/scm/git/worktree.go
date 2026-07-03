package git

import (
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
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

func RestoreWorktree(worktreePath string, source string, staging bool) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	gitArgs = append(gitArgs, "restore", "--source="+source, "--worktree")
	if staging {
		gitArgs = append(gitArgs, "--staged")
	}
	// ":/" is git's magic pathspec for the repository root, so every tracked
	// file is restored regardless of the current working directory.
	err := seedshell.ImpureRun("git", append(gitArgs, "--", ":/")...)
	if err != nil {
		return seederr.WrapErrorf("failed to restore from %v: %w", source, err)
	}
	return nil
}

func CreateCommit(worktreePath string, message string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	statuses, err := GetStatus(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(statuses) == 0 {
		return seederr.WrapErrorf("no changes to commit")
	}
	// In git's porcelain status the first character is the staging-area (index)
	// state and the second is the worktree state; "??" marks an untracked file.
	staged := 0
	unstaged := 0
	for _, s := range statuses {
		if s.Status == "??" {
			unstaged++
			continue
		}
		if s.Status[0] != ' ' {
			staged++
		}
		if s.Status[1] != ' ' {
			unstaged++
		}
	}
	if staged == 0 {
		seedlog.Warnf("Staging all change(s) first. count=%v", unstaged)
		err = seedshell.ImpureRun("git", append(gitArgs, "add", "--all")...)
		if err != nil {
			return seederr.WrapErrorf("failed to stage all changes: %w", err)
		}
	} else if unstaged > 0 {
		seedlog.Warnf("Unstaged change(s) exist. count=%v", unstaged)
	}
	err = seedshell.ImpureRun("git", append(gitArgs, "commit", "-m", message)...)
	if err != nil {
		return seederr.WrapErrorf("failed to commit: %w", err)
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
