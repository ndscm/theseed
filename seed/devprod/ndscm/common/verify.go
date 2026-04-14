package common

import (
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func QuickVerifyGitMonorepo(ndConfig *NdConfig) error {
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	if monorepoHome == "" {
		return seederr.WrapErrorf("monorepo home is not defined")
	}
	monorepoHomeStat, err := os.Stat(monorepoHome)
	if os.IsNotExist(err) {
		return seederr.WrapErrorf("monorepo home (%v) does not exist", monorepoHome)
	}
	if err != nil {
		return seederr.Wrap(err)
	}
	if !monorepoHomeStat.IsDir() {
		return seederr.WrapErrorf("monorepo home (%v) is not a folder", monorepoHome)
	}

	scmName, err := scm.ScmName()
	if err != nil {
		return seederr.Wrap(err)
	}
	switch scmName {
	case "git":
		// pass
	default:
		return seederr.WrapErrorf("scm is unsupported: %v", scmName)
	}

	monorepoGitDir, err := scm.MonorepoGitDir()
	if err != nil {
		return seederr.Wrap(err)
	}
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
