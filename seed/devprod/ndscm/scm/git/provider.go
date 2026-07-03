package git

import (
	"errors"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func sanitizeTrailerKey(key string) (string, error) {
	key = strings.ToLower(strings.TrimSpace(key))
	if !trailerKeyRegex.MatchString(key) {
		return "", seederr.WrapErrorf("invalid trailer key")
	}
	key = strings.ToUpper(key[:1]) + key[1:]
	return key, nil
}

type GitProvider struct{}

func (g *GitProvider) Initialize() error {
	return nil
}

// # amend

func (g *GitProvider) AmendAppendExtendedMetadata(key string, value string) error {
	trailerKey, err := sanitizeTrailerKey(key)
	if err != nil {
		return seederr.Wrap(err)
	}
	return AmendAppendTrailer("", trailerKey, value)
}

// # connect

func (g *GitProvider) Connect(
	repoIdentifier string, monorepoHome string, repoEndpoint string,
) (string, string, error) {
	repoEnvLines := []string{}
	repoEnvLines = append(repoEnvLines, `ND_MONOREPO_HOME="`+monorepoHome+`"`)

	userHandle, err := user.CurrentUserHandle()
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	repoEnvLines = append(repoEnvLines, `ND_USER_HANDLE="`+userHandle+`"`)
	userEmail, err := user.CurrentUserEmail()
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	repoEnvLines = append(repoEnvLines, `ND_USER_EMAIL="`+userEmail+`"`)
	userDisplayName, err := user.CurrentUserDisplayName()
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	repoEnvLines = append(repoEnvLines, `ND_USER_DISPLAY_NAME="`+userDisplayName+`"`)

	repoEnvLines = append(repoEnvLines, `ND_SCM="git"`)

	gitDir := guessMonorepoGitDir(monorepoHome)

	mainBranch, err := BareClone(repoEndpoint, gitDir)
	if err != nil {
		return "", "", seederr.Wrap(err)
	}
	remoteMainBranch := "origin/" + mainBranch
	// Sometimes git automatically creates the local main branch
	_, err = GetBranch(gitDir, mainBranch)
	if err != nil {
		if errors.Is(err, scm.ErrBranchNotFound) {
			err = CreateBranch(gitDir, mainBranch, remoteMainBranch, remoteMainBranch)
			if err != nil {
				return "", "", seederr.Wrap(err)
			}
		} else {
			return "", "", seederr.Wrap(err)
		}
	}
	worktreePath, err := CreateBranchWorktree(gitDir, monorepoHome, mainBranch)
	if err != nil {
		return "", "", seederr.Wrap(err)
	}

	err = os.WriteFile(filepath.Join(monorepoHome, "ndscm.env"), []byte(strings.Join(repoEnvLines, "\n")+"\n"), 0644)
	if err != nil {
		return "", "", seederr.WrapErrorf("failed to create ndscm.env file: %w", err)
	}
	err = SetConfig(gitDir, "user.name", userDisplayName)
	if err != nil {
		return "", "", seederr.WrapErrorf("failed to set git user.name: %w", err)
	}
	err = SetConfig(gitDir, "user.email", userEmail)
	if err != nil {
		return "", "", seederr.WrapErrorf("failed to set git user.email: %w", err)
	}
	return remoteMainBranch, worktreePath, nil
}

// # verify

func (g *GitProvider) QuickVerifyMonorepo() error {
	return QuickVerifyMonorepo()
}

// # branch

func (g *GitProvider) CreateBranch(branchName string, startPoint string, tracking string) error {
	return CreateBranch("", branchName, startPoint, tracking)
}

func (g *GitProvider) GetBranch(branchName string) (string, error) {
	return GetBranch("", branchName)
}

func (g *GitProvider) UpdateBranch(branchName string, newPoint string) error {
	return UpdateBranch("", branchName, newPoint)
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

func (g *GitProvider) GetCommitMetadata(commit string) (*scm.CommitMetadata, error) {
	metadata, err := GetCommitMetadata("", commit)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	result := &scm.CommitMetadata{
		CommitId:       metadata.Hash,
		Author:         metadata.Author,
		AuthorEmail:    metadata.AuthorEmail,
		AuthorTime:     metadata.AuthorTime,
		Committer:      metadata.Committer,
		CommitterEmail: metadata.CommitterEmail,
		CommitterTime:  metadata.CommitterTime,
		Subject:        metadata.Subject,
		Body:           metadata.Body,
	}
	for _, trailer := range metadata.Trailers {
		trailerKey := strings.ToLower(trailer.Key)
		switch trailerKey {
		case "change-uuid":
			if result.ChangeUuid != "" {
				return nil, seederr.WrapErrorf("multiple change-uuid found. commit=%v", metadata.Hash)
			}
			result.ChangeUuid = trailer.Value
		default:
			result.Extended = append(result.Extended, scm.KeyValue{
				Key:   trailerKey,
				Value: trailer.Value,
			})
		}
	}
	return result, nil
}

func (g *GitProvider) GetCommitFormatPatch(commit string) (string, error) {
	return GetFormatPatch("", commit)
}

// # filetree

func (g *GitProvider) ListCommitFiles(commit string) ([]scm.FileStatus, error) {
	result := []scm.FileStatus{}
	files, err := ListCommitFiles("", commit)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	for _, f := range files {
		result = append(result, scm.FileStatus{
			Status: f.Status,
			To:     f.To,
			From:   f.From,
		})
	}
	return result, nil
}

// # rebase

func (g *GitProvider) ApplyFormatPatch(worktreePath string, formatPatch string) error {
	return ApplyFormatPatch(worktreePath, formatPatch)
}

func (g *GitProvider) Rebase(worktreePath string, upstream string) error {
	return Rebase(worktreePath, upstream)
}

func (g *GitProvider) PullRebase(worktreePath string) error {
	return PullRebase(worktreePath)
}

func (g *GitProvider) ApplyCommitRange(worktreePath string, from string, to string) error {
	return CherryPickRange(worktreePath, from, to)
}

func (g *GitProvider) SignOff(worktreePath string) error {
	return SignOff(worktreePath)
}

// # remote

func (g *GitProvider) FetchAll() error {
	return FetchAll("")
}

func (g *GitProvider) FetchBranch(remote string, remoteBranchName string, branchName string) error {
	return FetchBranch("", remote, remoteBranchName, branchName)
}

func (g *GitProvider) PushBranch(branchName string, remote string, remoteBranchName string) error {
	return PushBranch("", branchName, remote, remoteBranchName)
}

func (g *GitProvider) ListRemoteBranches(remote string) ([]string, error) {
	return ListRemoteBranches("", remote)
}

// # search

func (g *GitProvider) SearchForkPoint(ourTipPoint string, theirTipPoint string) (string, string, error) {
	return SearchForkPoint("", ourTipPoint, theirTipPoint)
}

func (g *GitProvider) SearchExtendedMetadata(tipPoint string, key string, value string) (string, error) {
	trailerKey, err := sanitizeTrailerKey(key)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return SearchTrailer("", tipPoint, trailerKey, value)
}

// # status

func (g *GitProvider) ListDirtyFiles(worktreePath string) ([]scm.FileStatus, error) {
	files, err := GetStatus(worktreePath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	result := []scm.FileStatus{}
	for _, s := range files {
		result = append(result, scm.FileStatus{
			Status: s.Status,
			To:     s.To,
			From:   s.From,
		})
	}
	return result, nil
}

func (g *GitProvider) ListFiles(worktreePath string) ([]string, error) {
	return ListFiles(worktreePath, true)
}

func (g *GitProvider) GetWorktreeBranch(worktreePath string) (string, error) {
	return GetCurrentBranch(worktreePath)
}

func (g *GitProvider) GetWorktreeOperation(worktreePath string) (string, error) {
	return GetCurrentOperation(worktreePath)
}

// # worktree

func (g *GitProvider) GetCurrentWorktree(monorepoHome string) (string, string, error) {
	worktreePath, err := GetCurrentWorktreePath()
	if err != nil {
		return "", "", seederr.Wrap(err)
	}

	// current worktree may not be connected with ndscm (e.g. ci environment)
	worktreeName := ""
	if monorepoHome != "" {
		tmpWorktreeName, err := filepath.Rel(monorepoHome, worktreePath)
		if err != nil {
			return "", "", seederr.Wrap(err)
		}
		worktreeName = tmpWorktreeName
	}

	return worktreeName, worktreePath, nil
}

func (g *GitProvider) Checkout(worktreePath string, branchName string) error {
	return Checkout(worktreePath, branchName)
}

func (g *GitProvider) CreateCommit(worktreePath string, message string) error {
	return CreateCommit(worktreePath, message)
}

func (g *GitProvider) CreateWorktree(monorepoHome string, worktreeName string) (string, error) {
	monorepoGitDir := guessMonorepoGitDir(monorepoHome)
	return CreateBranchWorktree(monorepoGitDir, monorepoHome, worktreeName)
}

func (g *GitProvider) RemoveWorktree(monorepoHome string, worktreeName string) error {
	monorepoGitDir := guessMonorepoGitDir(monorepoHome)
	worktreePath := filepath.Join(monorepoHome, worktreeName)
	return RemoveWorktree(monorepoGitDir, worktreePath)
}

var _ scm.Provider = (*GitProvider)(nil)
