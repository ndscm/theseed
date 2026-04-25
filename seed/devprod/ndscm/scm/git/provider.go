package git

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
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

// # branch

func (g *GitProvider) CreateBranch(branchName string, startPoint string, tracking string) error {
	return CreateBranch("", branchName, startPoint, tracking)
}

func (g *GitProvider) GetBranchTracking(branchName string) (string, error) {
	return GetBranchTracking("", branchName)
}

func (g *GitProvider) SetBranchTracking(branchName string, tracking string) error {
	return SetBranchTracking("", branchName, tracking)
}

// # commit

func (g *GitProvider) GetCommitId(commit string) (string, error) {
	return GetCommitHash("", commit)
}

func (g *GitProvider) ListCommitIds(from string, to string) ([]string, error) {
	mergeCommits, err := ListMergeCommitHash("", from, to)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if len(mergeCommits) > 0 {
		return nil, seederr.WrapErrorf("current segment (%v..%v) is not pure and contains merge commit:\n%v", from, to, mergeCommits)
	}
	return ListCommitHash("", from, to)
}

// # status

func (g *GitProvider) GetWorktreeDirtyFiles(worktreePath string) ([]scm.DirtyFile, error) {
	files, err := GetStatus(worktreePath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	result := []scm.DirtyFile{}
	for _, s := range files {
		result = append(result, scm.DirtyFile{
			Status: s.Status,
			To:     s.To,
			From:   s.From,
		})
	}
	return result, nil
}

func (g *GitProvider) GetWorktreeBranch(worktreePath string) (string, error) {
	return GetCurrentBranch(worktreePath)
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
