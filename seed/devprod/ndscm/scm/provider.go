package scm

import (
	"time"
)

// FileStatus describes a single modified or untracked path in a worktree.
type FileStatus struct {
	// TODO(nagi): add staging status
	Status string
	To     string
	From   string
}

type KeyValue struct {
	Key   string
	Value string
}

type CommitMetadata struct {
	CommitId       string
	ChangeUuid     string
	Author         string
	AuthorEmail    string
	AuthorTime     time.Time
	Committer      string
	CommitterEmail string
	CommitterTime  time.Time
	Subject        string
	Body           string

	// Ordered list of other trailers, which may contain duplicates.
	Extended []KeyValue
}

type WipSplitStatus struct {
	Belong string `json:"belong"`

	CommitId string `json:"commitId"`
}

type WipStatus struct {
	Operation string `json:"operation"`

	Split *WipSplitStatus `json:"split,omitempty"`
}

// Provider is the abstraction over a Source Code Management backend (e.g.
// git). Each backend registers a provider via Register, and the active one is
// selected at runtime through the --scm flag.
//
// All worktreePath parameters accept "" to mean the current working directory.
type Provider interface {
	// Initialize prepares the provider for use. It is called exactly once,
	// before any other method.
	Initialize() error

	// # amend

	// AmendAppendExtendedMetadata amends the current worktree's head commit,
	// appending a "<key>: <value>" extended metadata entry to its message
	// without otherwise editing it.
	AmendAppendExtendedMetadata(key string, value string) error

	// # connect

	// Connect clones repoEndpoint into monorepoHome, sets up the default
	// branch, and creates a worktree for it. It returns the remote default
	// branch (e.g. "origin/main") and the worktree path.
	Connect(
		repoIdentifier string, monorepoHome string, repoEndpoint string,
	) (remoteDefaultBranch string, worktreePath string, err error)

	// # verify

	// QuickVerifyMonorepo performs a fast sanity check that the current
	// working directory belongs to a valid monorepo for this SCM.
	QuickVerifyMonorepo() error

	// # branch

	// CreateBranch creates branchName at startPoint. If tracking is non-empty,
	// the new branch is configured to track it.
	CreateBranch(branchName string, startPoint string, tracking string) error

	// CreateOrphanBranch creates branchName as an orphan branch with no history,
	// rooted at a single empty commit carrying message.
	CreateOrphanBranch(branchName string, message string) error

	// GetBranch resolves branchName to its commit id.
	GetBranch(branchName string) (string, error)

	// UpdateBranch force-moves branchName to point at newPoint, recording a
	// reflog entry for the move.
	UpdateBranch(branchName string, newPoint string) error

	// DeleteBranch removes branchName even if it has unmerged commits.
	DeleteBranch(branchName string) error

	// DeleteMergedBranch removes branchName only if its commits are reachable
	// from its upstream; it fails for unmerged branches.
	DeleteMergedBranch(branchName string) error

	// GetBranchTracking returns the upstream branch that branchName tracks.
	GetBranchTracking(branchName string) (string, error)

	// SetBranchTracking configures branchName to track the given upstream.
	SetBranchTracking(branchName string, tracking string) error

	// GetMergeBaseCommitId returns the commit id of the most recent common
	// ancestor of base and target.
	GetMergeBaseCommitId(base string, target string) (string, error)

	// # commit

	// GetCommitId resolves commit (a ref, partial id, or full id) to a full
	// commit id.
	GetCommitId(commit string) (string, error)

	// ListCommitIds returns the commit ids in the linear range (from, to].
	// It fails if the segment is not pure (e.g. contains a merge commit).
	ListCommitIds(from string, to string) ([]string, error)

	// GetCommitMetadata resolves commit and returns its metadata, including
	// authorship, commit and author times, the subject and body, and any
	// extended trailers.
	GetCommitMetadata(commit string) (*CommitMetadata, error)

	// GetCommitFormatPatch renders commit as a single format-patch (header,
	// message, and per-file diffs).
	GetCommitFormatPatch(commit string) (string, error)

	// # filetree

	// ListCommitFiles lists every committed files in single commit.
	ListCommitFiles(commit string) ([]FileStatus, error)

	// # rebase

	// ApplyFormatPatch applies patch as a new commit on top of the worktree's
	// current branch, preserving the original commit message and trailers.
	ApplyFormatPatch(worktreePath string, formatPatch string) error

	// Rebase replays the commits of the worktree's current branch that are not
	// in upstream on top of onto. If onto is empty, it replays them on top of
	// upstream.
	Rebase(worktreePath string, upstream string, onto string) error

	// PullRebase fetches the worktree branch's tracking remote and rebases
	// onto the updated tracking branch.
	PullRebase(worktreePath string) error

	// ApplyCommitRange replays the commits in (from, to] onto the worktree's
	// current branch.
	ApplyCommitRange(worktreePath string, from string, to string) error

	// SignOff rebases the worktree's current branch with a sign-off.
	SignOff(worktreePath string) error

	// # remote

	// FetchAll fetches updates from every configured remote and prunes
	// remote-tracking refs that no longer exist upstream.
	FetchAll() error

	// FetchBranch fetches remoteBranchName from remote and stores it as
	// the local ref branchName. Stale remote-tracking refs are pruned.
	FetchBranch(remote string, remoteBranchName string, branchName string) error

	// PushBranch force-pushes the local branchName to remote as
	// remoteBranchName.
	PushBranch(branchName string, remote string, remoteBranchName string) error

	// ListRemoteBranches returns the short names of remote-tracking branches
	// under remote.
	ListRemoteBranches(remote string) ([]string, error)

	// # search

	// SearchForkPoint walks both commit chains backwards from ourTipPoint and
	// theirTipPoint by committer time, matching Change-uuid trailers, and
	// returns the pair of commits where the two chains first share a
	// common Change-uuid. It fails if either chain contains a merge
	// commit or if both chains reach their root without finding a match.
	SearchForkPoint(ourTipPoint string, theirTipPoint string) (string, string, error)

	// SearchExtendedMetadata searches for a commit with a specific extended
	// metadata key and value. If value is empty, it searches for the presence
	// of the key regardless of its value. It returns the commit id of the most
	// recent matching commit, or an error if no match is found.
	SearchExtendedMetadata(tipPoint string, key string, value string) (string, error)

	// # status

	// ListDirtyFiles lists every modified or untracked file in worktreePath.
	ListDirtyFiles(worktreePath string) ([]FileStatus, error)

	// ListFiles lists every tracked and untracked files in worktreePath.
	ListFiles(worktreePath string) ([]string, error)

	// GetWorktreeBranch returns the branch currently checked out in
	// worktreePath.
	GetWorktreeBranch(worktreePath string) (string, error)

	// GetWorktreeOperation returns the name of any in-progress operation
	// (e.g. "rebase", "merge", "cherry-pick") in worktreePath, or "" if
	// idle.
	GetWorktreeOperation(worktreePath string) (string, error)

	// # wip

	// LoadWipStatus returns the wip status persisted for worktreePath, or nil
	// when no wip operation is in progress.
	LoadWipStatus(worktreePath string) (*WipStatus, error)

	// SaveWipStatus persists status for worktreePath, overwriting any existing
	// wip status.
	SaveWipStatus(worktreePath string, status *WipStatus) error

	// RemoveWipStatus deletes the wip status persisted for worktreePath. When
	// force is true a missing status is not an error.
	RemoveWipStatus(worktreePath string, force bool) error

	// # worktree

	// GetCurrentWorktree returns the worktree containing the current working
	// directory, as both its name (the worktree path relative to
	// monorepoHome) and its absolute path. It fails if the current worktree
	// is not under monorepoHome.
	GetCurrentWorktree(monorepoHome string) (worktreeName string, worktreePath string, err error)

	// Checkout switches worktreePath to branchName.
	Checkout(worktreePath string, branchName string) error

	// RestoreWorktree overwrites every tracked file in worktreePath with its
	// content at source, leaving the checked-out branch unchanged. When staging
	// is true it also overwrites the staging area (index) so that both the
	// worktree files and the index match source.
	RestoreWorktree(worktreePath string, source string, staging bool) error

	// UpdateStagingArea moves path — a file or directory — into the staging area
	// when staging is true, or out of it (unstages) when staging is false.
	// Staging never force-adds git-ignored content, and neither direction is an
	// error when no files are actually moved.
	UpdateStagingArea(worktreePath string, path string, staging bool) error

	// CreateCommit records the staged changes in worktreePath as a new commit
	// with the given message. When nothing is staged it first stages every
	// change, including untracked files; when a subset is already staged it
	// commits only that subset and leaves the rest unstaged. It fails if there
	// are no changes to commit.
	CreateCommit(worktreePath string, message string) error

	// CreateCommitReuse records a new commit from the staged content of
	// worktreePath, reusing the message and authorship of commit without
	// editing them. It always creates a commit, even when the staged content is
	// identical to the parent.
	CreateCommitReuse(worktreePath string, commit string) error

	// CreateWorktree materializes the worktree at the conventional path for
	// worktreeName under monorepoHome and returns that path. A branch named
	// worktreeName must already exist; it is the branch checked out in the
	// new worktree.
	CreateWorktree(monorepoHome string, worktreeName string) (string, error)

	// RemoveWorktree deletes the worktree for worktreeName under
	// monorepoHome. The branch of the same name is left intact.
	RemoveWorktree(monorepoHome string, worktreeName string) error
}
