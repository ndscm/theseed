package scm

// scm.Provider is the abstraction over a Source Code Management backend (e.g.
// git). Each backend registers a provider via Register, and the active one is
// selected at runtime through the --scm flag.
type Provider interface {
	// Initialize prepares the provider for use. It is called exactly once,
	// before any other method.
	Initialize() error

	// # verify

	// QuickVerifyMonorepo performs a fast sanity check that the current
	// working directory belongs to a valid monorepo for this SCM.
	QuickVerifyMonorepo() error

	// # worktree

	// GetBranchWorktree returns the conventional worktree path for branchName
	// under monorepoHome. It does not check whether the worktree exists.
	GetBranchWorktree(monorepoHome string, branchName string) string

	// CreateDevWorktree creates the worktree pair for a dev branch under
	// monorepoHome: a base/<branchName> branch tracking origin/main, and the
	// dev branchName tracking that base. Returns the new worktree path.
	CreateDevWorktree(monorepoHome string, branchName string) (string, error)
}
