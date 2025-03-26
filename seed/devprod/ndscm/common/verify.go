package common

import (
	"fmt"
	"os"
)

func QuickVerifyGitMonorepo(ndConfig *NdConfig) error {
	if ndConfig.MonorepoHome == "" {
		return WrapTrace(fmt.Errorf("monorepo home (ND_MONOREPO_HOME) is not defined"))
	}
	monorepoHomeStat, err := os.Stat(ndConfig.MonorepoHome)
	if os.IsNotExist(err) {
		return WrapTrace(fmt.Errorf("monorepo home (%v) does not exist", ndConfig.MonorepoHome))
	}
	if err != nil {
		return WrapTrace(err)
	}
	if !monorepoHomeStat.IsDir() {
		return WrapTrace(fmt.Errorf("monorepo home (%v) is not a folder", ndConfig.MonorepoHome))
	}
	if ndConfig.Scm != "git" {
		return WrapTrace(fmt.Errorf("SCM %v is not supported", ndConfig.Scm))
	}
	if ndConfig.MonorepoGitDir == "" {
		return WrapTrace(fmt.Errorf("monorepo git dir (ND_MONOREPO_GIT_DIR) is not defined"))
	}
	monorepoGitDirStat, err := os.Stat(ndConfig.MonorepoGitDir)
	if os.IsNotExist(err) {
		return WrapTrace(fmt.Errorf("monorepo git dir (%v) does not exist", ndConfig.MonorepoGitDir))
	}
	if err != nil {
		return WrapTrace(err)
	}
	if !monorepoGitDirStat.IsDir() {
		return WrapTrace(fmt.Errorf("monorepo git dir (%v) is not a folder", ndConfig.MonorepoGitDir))
	}
	return nil
}
