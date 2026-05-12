package runner

import (
	"slices"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type Runner struct {
	worktreePath string

	scmFilePaths []string

	scmChangePaths []string

	scmDirtyPaths []string
}

func (r *Runner) Format(all bool) error {
	if !all {
		err := FormatDirtyFiles(r.worktreePath, r.scmFilePaths, slices.Concat(r.scmChangePaths, r.scmDirtyPaths))
		if err != nil {
			return seederr.Wrap(err)
		}
	} else {
		err := FormatAllFiles(r.worktreePath, r.scmFilePaths)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

func (r *Runner) GenerateMakefile() (string, error) {
	return GenerateMakefile(r.worktreePath, r.scmFilePaths)
}

func CreateRunner(
	worktreePath string, scmFilePaths []string, scmChangePaths []string, scmDirtyPaths []string,
) (*Runner, error) {
	r := &Runner{
		worktreePath:   worktreePath,
		scmFilePaths:   scmFilePaths,
		scmChangePaths: scmChangePaths,
		scmDirtyPaths:  scmDirtyPaths,
	}
	return r, nil
}
