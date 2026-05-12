package clientcore

import (
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/ndscm/runner"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type NdMakefileOptions struct {
}

func NdMakefile(scmProvider scm.Provider, _ NdMakefileOptions) error {
	worktreePath, err := scmProvider.GetCurrentWorktree()
	if err != nil {
		return seederr.Wrap(err)
	}
	filePaths, err := scmProvider.ListFiles(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	changePaths, err := scmProvider.ListCommitFiles("HEAD")
	if err != nil {
		return seederr.Wrap(err)
	}
	dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	dirtyPaths := []string{}
	for _, dirtyFile := range dirtyFiles {
		dirtyPaths = append(dirtyPaths, dirtyFile.To)
	}
	r, err := runner.CreateRunner(worktreePath, filePaths, changePaths, dirtyPaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	makefile, err := r.GenerateMakefile()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(filepath.Join(worktreePath, "ndscm.Makefile"), []byte(makefile), 0644)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
