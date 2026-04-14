package common

import (
	"os"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func QuickVerifyGitMonorepo(ndConfig *NdConfig) error {
	if ndConfig.MonorepoHome == "" {
		return seederr.WrapErrorf("monorepo home (ND_MONOREPO_HOME) is not defined")
	}
	monorepoHomeStat, err := os.Stat(ndConfig.MonorepoHome)
	if os.IsNotExist(err) {
		return seederr.WrapErrorf("monorepo home (%v) does not exist", ndConfig.MonorepoHome)
	}
	if err != nil {
		return seederr.Wrap(err)
	}
	if !monorepoHomeStat.IsDir() {
		return seederr.WrapErrorf("monorepo home (%v) is not a folder", ndConfig.MonorepoHome)
	}
	if ndConfig.Scm != "git" {
		return seederr.WrapErrorf("SCM %v is not supported", ndConfig.Scm)
	}
	if ndConfig.MonorepoGitDir == "" {
		return seederr.WrapErrorf("monorepo git dir (ND_MONOREPO_GIT_DIR) is not defined")
	}
	monorepoGitDirStat, err := os.Stat(ndConfig.MonorepoGitDir)
	if os.IsNotExist(err) {
		return seederr.WrapErrorf("monorepo git dir (%v) does not exist", ndConfig.MonorepoGitDir)
	}
	if err != nil {
		return seederr.Wrap(err)
	}
	if !monorepoGitDirStat.IsDir() {
		return seederr.WrapErrorf("monorepo git dir (%v) is not a folder", ndConfig.MonorepoGitDir)
	}
	return nil
}
