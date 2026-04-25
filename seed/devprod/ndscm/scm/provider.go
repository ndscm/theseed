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

	// GetBranchTracking returns the upstream branch that branchName tracks.
	GetBranchTracking(branchName string) (string, error)

	// SetBranchTracking configures branchName to track tracking.
	SetBranchTracking(branchName string, tracking string) error

	// # commit

	// GetCommitId resolves commit (a ref, partial id, or full id) to a full
	// commit id.
	GetCommitId(commit string) (string, error)

	// ListCommitIds returns the commit ids in the linear range (from, to].
	// It fails if the segment is not pure (e.g. contains a merge commit).
	ListCommitIds(from string, to string) ([]string, error)

	// # status

	// GetWorktreeDirtyFiles lists every modified or untracked file in
	// worktreePath.
	GetWorktreeDirtyFiles(worktreePath string) ([]DirtyFile, error)

	// GetWorktreeBranch returns the branch currently checked out in
	// worktreePath.
	GetWorktreeBranch(worktreePath string) (string, error)

	// # worktree

	// GetBranchWorktree returns the conventional worktree path for branchName
	// under monorepoHome. It does not check whether the worktree exists.
	GetBranchWorktree(monorepoHome string, branchName string) string

	// CreateDevWorktree creates the worktree pair for a dev branch under
	// monorepoHome: a base/<branchName> branch tracking origin/main, and the
	// dev branchName tracking that base. Returns the new worktree path.
	CreateDevWorktree(monorepoHome string, branchName string) (string, error)
}
