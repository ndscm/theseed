package git

import (
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type GitProvider struct{}

func (g *GitProvider) Initialize() error {
	return nil
}

// # verify

func (g *GitProvider) QuickVerifyMonorepo() error {
	return QuickVerifyMonorepo()
}

// # worktree

func (g *GitProvider) GetBranchWorktree(monorepoHome string, branchName string) string {
	return GetBranchWorktreePath(monorepoHome, branchName)
}

func (g *GitProvider) CreateDevWorktree(monorepoHome string, worktreeName string) (string, error) {
	monorepoGitDir, err := MonorepoGitDir()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = CreateBranch(monorepoGitDir, "base/"+worktreeName, "origin/main", "origin/main")
	if err != nil {
		return "", seederr.WrapErrorf("failed to create base branch %v: %v", "base/"+worktreeName, err)
	}
	err = CreateBranch(monorepoGitDir, worktreeName, "base/"+worktreeName, "base/"+worktreeName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create worktree branch %v: %v", worktreeName, err)
	}
	newWorktreePath, err := CreateBranchWorktree(monorepoGitDir, monorepoHome, worktreeName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create branch worktree %v: %v", worktreeName, err)
	}
	return newWorktreePath, nil
}
