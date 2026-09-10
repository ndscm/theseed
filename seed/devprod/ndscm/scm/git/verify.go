package git

import (
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func QuickVerifyMonorepo(repo *scm.WorkingRepo) error {
	if repo == nil || repo.MonorepoHome == "" {
		return seederr.WrapErrorf("working repo is not connected by ndscm")
	}
	monorepoHomeStat, err := os.Stat(repo.MonorepoHome)
	if os.IsNotExist(err) {
		return seederr.WrapErrorf("monorepo home (%v) does not exist", repo.MonorepoHome)
	}
	if err != nil {
		return seederr.Wrap(err)
	}
	if !monorepoHomeStat.IsDir() {
		return seederr.WrapErrorf("monorepo home (%v) is not a folder", repo.MonorepoHome)
	}

	monorepoGitDir := guessMonorepoGitDir(repo)
	if monorepoGitDir == "" {
		return seederr.WrapErrorf("monorepo git dir is not defined")
	}
	monorepoGitDirStat, err := os.Stat(monorepoGitDir)
	if os.IsNotExist(err) {
		return seederr.WrapErrorf("monorepo git dir (%v) does not exist", monorepoGitDir)
	}
	if err != nil {
		return seederr.Wrap(err)
	}
	if !monorepoGitDirStat.IsDir() {
		return seederr.WrapErrorf("monorepo git dir (%v) is not a folder", monorepoGitDir)
	}
	return nil
}
