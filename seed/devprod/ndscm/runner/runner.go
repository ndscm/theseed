package runner

import (
	"slices"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type Runner struct {
	worktree string

	scmFilePaths []string

	scmChangePaths []string

	scmDirtyPaths []string
}

func (r *Runner) Format(all bool) error {
	if !all {
		err := FormatDirtyFiles(r.worktree, r.scmFilePaths, slices.Concat(r.scmChangePaths, r.scmDirtyPaths))
		if err != nil {
			return seederr.Wrap(err)
		}
	} else {
		err := FormatAllFiles(r.worktree, r.scmFilePaths)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

func (r *Runner) GenerateMakefile() (string, error) {
	return GenerateMakefile(r.worktree, r.scmFilePaths)
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
