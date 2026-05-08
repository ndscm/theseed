package clientcore

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/runner"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdFormatOptions struct {
	All bool
}

func NdFormat(scmProvider scm.Provider, options NdFormatOptions) error {
	worktree, err := scmProvider.GetCurrentWorktree()
	if err != nil {
		return seederr.Wrap(err)
	}
	filePaths, err := scmProvider.ListFiles(worktree)
	if err != nil {
		return seederr.Wrap(err)
	}
	changePaths, err := scmProvider.ListCommitFiles("HEAD")
	if err != nil {
		return seederr.Wrap(err)
	}
	dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles(worktree)
	if err != nil {
		return seederr.Wrap(err)
	}
	dirtyPaths := []string{}
	for _, dirtyFile := range dirtyFiles {
		dirtyPaths = append(dirtyPaths, dirtyFile.To)
	}
	r, err := runner.CreateRunner(worktree, filePaths, changePaths, dirtyPaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = r.Format(options.All)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
