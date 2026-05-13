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
				dirtyPaths = append(dirtyPaths, repoPath)
			}
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
