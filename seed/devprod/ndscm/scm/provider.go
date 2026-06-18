package scm

// FileStatus describes a single modified or untracked path in a worktree.
type FileStatus struct {
	// TODO(nagi): add staging status
	Status string
	To     string
	From   string
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

	// # connect

	// Connect clones repoEndpoint into monorepoHome, sets up the default
	// branch, and creates a worktree for it. It returns the remote default
	// branch (e.g. "origin/main") and the worktree path.
	Connect(
		repoIdentifier string, monorepoHome string, repoEndpoint string, canonicalBranch bool,
	) (remoteDefaultBranch string, worktreePath string, err error)

	// # verify

	// QuickVerifyMonorepo performs a fast sanity check that the current
	// working directory belongs to a valid monorepo for this SCM.
	QuickVerifyMonorepo() error

	// # branch

	// CreateBranch creates branchName at startPoint. If tracking is non-empty,
	// the new branch is configured to track it.
	CreateBranch(branchName string, startPoint string, tracking string) error

	// GetBranch resolves branchName to its commit id.
	GetBranch(branchName string) (string, error)

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

	// IsDevBranch reports whether branchName is a dev branch. In canonical
	// mode that means an "<owner>/dev/<focus>" branch whose owner is not a
	// reserved handle; otherwise it means "dev" itself or a focused variant
	// "dev-<focus>" (no slashes allowed).
	IsDevBranch(branchName string, canonicalBranch bool) bool

	// IsMeltBranch reports whether branchName is a melt branch. In canonical
	// mode that means an "<owner>/melt/<upstream>" branch whose owner is not a
	// reserved handle; otherwise it means "melt" itself or a focused variant
	// "melt-<upstream>" (no slashes allowed).
	IsMeltBranch(branchName string, canonicalBranch bool) bool

	// # commit

	// GetCommitId resolves commit (a ref, partial id, or full id) to a full
	// commit id.
	GetCommitId(commit string) (string, error)

	// ListCommitIds returns the commit ids in the linear range (from, to].
	// It fails if the segment is not pure (e.g. contains a merge commit).
	ListCommitIds(from string, to string) ([]string, error)

	// # filetree

	// ListCommitFiles lists every committed files in single commit.
	ListCommitFiles(commit string) ([]FileStatus, error)

	// # rebase

	// Rebase replays the commits of the worktree's current branch on top of
	// upstream.
	Rebase(worktreePath string, upstream string) error

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

	// SearchForkPoint walks both commit chains backwards from ourHead and
	// theirHead by committer time, matching Change-uuid trailers, and
	// returns the pair of commits where the two chains first share a
	// common Change-uuid. It fails if either chain contains a merge
	// commit or if both chains reach their root without finding a match.
	SearchForkPoint(ourHead string, theirHead string) (string, string, error)

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

	// # worktree

	// GetCurrentWorktree returns the path of the worktree containing the
	// current working directory.
	GetCurrentWorktree() (string, error)

	// Checkout switches worktreePath to branchName.
	Checkout(worktreePath string, branchName string) error

	// CreateBranchWorktree materializes branchName as a worktree at the
	// conventional path under monorepoHome and returns that path.
	CreateBranchWorktree(monorepoHome string, branchName string) (string, error)

	// GetBranchWorktree returns the conventional worktree path for branchName
	// under monorepoHome. It does not check whether the worktree exists.
	GetBranchWorktree(monorepoHome string, branchName string) string

	// GetBranchWorktreeBranch returns the branch name implied by worktreePath
	// under monorepoHome (the inverse of GetBranchWorktree). It fails if
	// worktreePath is not under monorepoHome.
	GetBranchWorktreeBranch(monorepoHome string, worktreePath string) (string, error)

	// RemoveWorktree deletes the worktree at worktreePath.
	RemoveWorktree(monorepoHome string, worktreePath string) error

	// GetDevWorktree returns the dev branch name and its conventional
	// worktree path under monorepoHome for ownerHandle and focus, along
	// with whether that worktree currently exists. The branch is
	// "<owner>/dev/<focus>" in canonical mode (focus defaults to "main") or
	// "dev"/"dev-<focus>" otherwise.
	GetDevWorktree(
		monorepoHome string, ownerHandle string, focus string, canonicalBranch bool,
	) (string, string, bool)

	// CreateDevWorktree creates a dev worktree under monorepoHome. It sets
	// up a base tracking branch at tracking, creates the working branch on
	// top of it, and materializes the worktree. The working branch is
	// "<owner>/dev/<focus>" in canonical mode (focus defaults to "main") or
	// "dev"/"dev-<focus>" otherwise, and the base branch is its matching
	// base branch. Pass "" for focus to get the default dev branch.
	CreateDevWorktree(
		monorepoHome string, ownerHandle string, focus string, tracking string, canonicalBranch bool,
	) (string, error)

	// RemoveDevWorktree removes the dev worktree, its dev branch, and
	// its base branch under monorepoHome. It returns the new working
	// directory if it changed, or "" if it did not.
	RemoveDevWorktree(
		monorepoHome string, ownerHandle string, focus string, canonicalBranch bool,
	) (string, error)

	// GetMeltWorktree returns the melt branch name and its conventional
	// worktree path under monorepoHome for ownerHandle and upstreamName,
	// along with whether that worktree currently exists. The branch is
	// "<owner>/melt/<upstreamName>" in canonical mode (upstreamName
	// defaults to "default") or "melt"/"melt-<upstreamName>" otherwise.
	GetMeltWorktree(
		monorepoHome string, ownerHandle string, upstreamName string, canonicalBranch bool,
	) (string, string, bool)

	// CreateMeltWorktree creates a worktree for melting upstream changes.
	// It creates a base branch at fromPoint tracking the given tracking
	// ref, a working branch at toPoint tracking the base branch, and
	// materializes the worktree. The working branch is
	// "<owner>/melt/<upstreamName>" in canonical mode or
	// "melt-<upstreamName>" otherwise, and the base branch is its matching
	// base branch. After creation, the base branch is updated to forkPoint
	// so that a subsequent rebase replays only the commits between
	// forkPoint and toPoint.
	CreateMeltWorktree(
		monorepoHome string, ownerHandle string,
		upstreamName string, fromPoint string, toPoint string, tracking string, forkPoint string,
		canonicalBranch bool,
	) (string, error)

	// RemoveMeltWorktree removes the melt worktree, its working branch,
	// and its base branch under monorepoHome. It fails if the worktree
	// has dirty files. It returns the new working directory if the
	// caller was inside the removed worktree, or "" otherwise.
	RemoveMeltWorktree(
		monorepoHome string, ownerHandle string, upstreamName string, canonicalBranch bool,
	) (string, error)
}
