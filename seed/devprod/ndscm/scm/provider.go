package scm

// DirtyFile describes a single modified or untracked path in a worktree.
type DirtyFile struct {
	// TODO(nagi): add staging status
	Status string
	To     string
	From   string
}

// scm.Provider is the abstraction over a Source Code Management backend (e.g.
// git). Each backend registers a provider via Register, and the active one is
// selected at runtime through the --scm flag.
//
// All worktreePath parameters accept "" to mean the current working directory.
type Provider interface {
	// Initialize prepares the provider for use. It is called exactly once,
	// before any other method.
	Initialize() error

	// # verify

	// QuickVerifyMonorepo performs a fast sanity check that the current
	// working directory belongs to a valid monorepo for this SCM.
	QuickVerifyMonorepo() error

	// # branch

	// CreateBranch creates branchName at startPoint. If tracking is non-empty,
	// the new branch is configured to track it.
	CreateBranch(branchName string, startPoint string, tracking string) error

	GetBranch(branchName string) (string, error)

	// DeleteBranch removes branchName even if it has unmerged commits.
	DeleteBranch(branchName string) error

	// DeleteMergedBranch removes branchName only if its commits are reachable
	// from its upstream; it fails for unmerged branches.
	DeleteMergedBranch(branchName string) error

	// GetBranchTracking returns the upstream branch that branchName tracks.
	GetBranchTracking(branchName string) (string, error)

	// SetBranchTracking configures branchName to track tracking.
	SetBranchTracking(branchName string, tracking string) error

	// GetMergeBaseCommitId returns the commit id of the most recent common
	// ancestor of base and target.
	GetMergeBaseCommitId(base string, target string) (string, error)

	// IsDevBranch reports whether branchName is a dev branch — either "dev"
	// itself or a focused variant "dev-<focus>" (no slashes allowed).
	IsDevBranch(branchName string) bool

	// # commit

	// GetCommitId resolves commit (a ref, partial id, or full id) to a full
	// commit id.
	GetCommitId(commit string) (string, error)

	// ListCommitIds returns the commit ids in the linear range (from, to].
	// It fails if the segment is not pure (e.g. contains a merge commit).
	ListCommitIds(from string, to string) ([]string, error)

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

	// # remote

	// FetchAll fetches updates from every configured remote and prunes
	// remote-tracking refs that no longer exist upstream.
	FetchAll() error

	// PushBranch force-pushes the local branchName to remote as
	// remoteBranchName.
	PushBranch(branchName string, remote string, remoteBranchName string) error

	// # status

	// GetWorktreeDirtyFiles lists every modified or untracked file in
	// worktreePath.
	GetWorktreeDirtyFiles(worktreePath string) ([]DirtyFile, error)

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
	RemoveWorktree(worktreePath string) error

	// CreateDevWorktree creates a dev worktree under monorepoHome. It sets
	// up a base tracking branch (base/dev or base/dev-<focus>) at
	// tracking, creates the working branch on top of it, and
	// materializes the worktree. Pass "" for focus to get the plain "dev"
	// branch.
	CreateDevWorktree(monorepoHome string, focus string, tracking string) (string, error)

	// GetDevWorktree returns the conventional worktree path for the dev
	// branch (or dev-<focus>) under monorepoHome. It does not check
	// whether the worktree exists.
	GetDevWorktree(monorepoHome string, focus string) string

	// RemoveDevWorktree removes the dev worktree, its dev branch, and
	// its base branch under monorepoHome. It returns the new working
	// directory if it changed, or "" if it did not.
	RemoveDevWorktree(monorepoHome string, focus string) (string, error)
}
