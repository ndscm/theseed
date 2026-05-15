package clientcore

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/runner"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdRunOptions struct {
	Workers int
	All     bool
	Changed bool

	Phases []string
}

func NdRun(scmProvider scm.Provider, options NdRunOptions) error {
	worktreePath, err := scmProvider.GetCurrentWorktree()
	if err != nil {
		return seederr.Wrap(err)
	}
	filePaths, err := scmProvider.ListFiles(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	carePaths := []string{}
	if options.All {
		carePaths = filePaths
	} else if options.Changed {
		dirtySet := map[string]bool{}
		headChanges, err := scmProvider.ListCommitFiles("HEAD")
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, headFileStatus := range headChanges {
			dirtySet[headFileStatus.To] = true
		}
		dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles(worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, dirtyFileStatus := range dirtyFiles {
			dirtySet[dirtyFileStatus.From] = false
			dirtySet[dirtyFileStatus.To] = true
		}
		for repoPath, dirty := range dirtySet {
			if dirty {
				carePaths = append(carePaths, repoPath)
			}
		}
	} else {
		// Do not filter repo analysis by care paths.
	}
	r, err := runner.CreateRunner(options.Workers, worktreePath, filePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = r.Run(options.Phases, carePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
