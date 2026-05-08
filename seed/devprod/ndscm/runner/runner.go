package runner

type Runner struct {
	worktree string

	scmFilePaths []string

	scmChangePaths []string

	scmDirtyPaths []string
}

func CreateRunner(worktree string, scmFilePaths []string, scmChangePaths []string, scmDirtyPaths []string) (*Runner, error) {
	r := &Runner{
		worktree:       worktree,
		scmFilePaths:   scmFilePaths,
		scmChangePaths: scmChangePaths,
		scmDirtyPaths:  scmDirtyPaths,
	}
	return r, nil
}
