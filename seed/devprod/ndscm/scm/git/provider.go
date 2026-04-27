package git

import (
	"strings"

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

func (g *GitProvider) DeleteBranch(branchName string) error {
	return DeleteBranch("", branchName)
}

func (g *GitProvider) DeleteMergedBranch(branchName string) error {
	return DeleteMergedBranch("", branchName)
}

func (g *GitProvider) GetBranchTracking(branchName string) (string, error) {
	return GetBranchTracking("", branchName)
}

func (g *GitProvider) SetBranchTracking(branchName string, tracking string) error {
	return SetBranchTracking("", branchName, tracking)
}

func (g *GitProvider) GetMergeBaseCommitId(base string, target string) (string, error) {
	return GetMergeBaseHash("", base, target)
}

func (g *GitProvider) IsDevBranch(branchName string) bool {
	if branchName == "dev" {
		return true
	}
	if strings.HasPrefix(branchName, "dev-") && !strings.Contains(branchName, "/") {
		return true
	}
	return false
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

// # rebase

func (g *GitProvider) Rebase(worktreePath string, upstream string) error {
	return Rebase(worktreePath, upstream)
}

func (g *GitProvider) PullRebase(worktreePath string) error {
	return PullRebase(worktreePath)
}

func (g *GitProvider) ApplyCommitRange(worktreePath string, from string, to string) error {
	return CherryPickRange(worktreePath, from, to)
}

// # remote

func (g *GitProvider) FetchAll() error {
	return FetchAll("")
}

func (g *GitProvider) PushBranch(branchName string, remote string, remoteBranchName string) error {
	return PushBranch("", branchName, remote, remoteBranchName)
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

func (g *GitProvider) GetCurrentWorktree() (string, error) {
	return GetCurrentWorktreePath()
}

func (g *GitProvider) Checkout(worktreePath string, branchName string) error {
	return Checkout(worktreePath, branchName)
}

func (g *GitProvider) CreateBranchWorktree(monorepoHome string, branchName string) (string, error) {
	monorepoGitDir, err := MonorepoGitDir()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return CreateBranchWorktree(monorepoGitDir, monorepoHome, branchName)
}

func (g *GitProvider) GetBranchWorktree(monorepoHome string, branchName string) string {
	return GetBranchWorktreePath(monorepoHome, branchName)
}

func (g *GitProvider) GetBranchWorktreeBranch(monorepoHome string, worktreePath string) (string, error) {
	return GetBranchWorktreeBranch(monorepoHome, worktreePath)
}

func (g *GitProvider) RemoveWorktree(worktreePath string) error {
	monorepoGitDir, err := MonorepoGitDir()
	if err != nil {
		return seederr.Wrap(err)
	}
	return RemoveWorktree(monorepoGitDir, worktreePath)
}

func (g *GitProvider) CreateDevWorktree(monorepoHome string, focus string) (string, error) {
	branchName := "dev"
	if focus != "" {
		branchName = "dev-" + focus
	}
	monorepoGitDir, err := MonorepoGitDir()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = CreateBranch(monorepoGitDir, "base/"+branchName, "origin/main", "origin/main")
	if err != nil {
		return "", seederr.WrapErrorf("failed to create base branch %v: %v", "base/"+branchName, err)
	}
	err = CreateBranch(monorepoGitDir, branchName, "base/"+branchName, "base/"+branchName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create worktree branch %v: %v", branchName, err)
	}
	newWorktreePath, err := CreateBranchWorktree(monorepoGitDir, monorepoHome, branchName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create branch worktree %v: %v", branchName, err)
	}
	return newWorktreePath, nil
}

func (g *GitProvider) GetDevWorktree(monorepoHome string, focus string) string {
	branchName := "dev"
	if focus != "" {
		branchName = "dev-" + focus
	}
	return GetBranchWorktreePath(monorepoHome, branchName)
}
