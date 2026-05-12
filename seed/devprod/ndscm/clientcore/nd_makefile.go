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
	makefile, err := runner.GenerateMakefile(worktreePath, filePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.MkdirAll(filepath.Join(worktreePath, ".cache/ndscm"), 0755)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(filepath.Join(worktreePath, ".cache/ndscm/Makefile"), []byte(makefile), 0644)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
