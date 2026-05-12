package clientcore

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/runner"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdFormatOptions struct {
	All     bool
	Changed bool
}

func NdFormat(scmProvider scm.Provider, options NdFormatOptions) error {
	worktreePath, err := scmProvider.GetCurrentWorktree()
	if err != nil {
		return seederr.Wrap(err)
	}
	filePaths, err := scmProvider.ListFiles(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	dirtyPaths := []string{}
	if options.All {
		dirtyPaths = filePaths
	} else if options.Changed {
		changePaths, err := scmProvider.ListCommitFiles("HEAD")
		if err != nil {
			return seederr.Wrap(err)
		}
		dirtyPaths = append(dirtyPaths, changePaths...)
		dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles(worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, dirtyFile := range dirtyFiles {
			dirtyPaths = append(dirtyPaths, dirtyFile.To)
		}
	} else {
		// Allow runner to determine dirty paths from stamps
	}
	r, err := runner.CreateRunner(worktreePath, filePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = r.Format(dirtyPaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
