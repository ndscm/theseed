package common

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagScm = seedflag.DefineString("scm", "git", "the scm backend")

var flagMonorepoHome = seedflag.DefineString("monorepo_home", "", "the monorepo home directory")
var flagMonorepoGitDir = seedflag.DefineString("monorepo_git_dir", "", "the monorepo git directory")
var flagUserHandle = seedflag.DefineString("user_handle", "", "the user handle")

type NdConfig struct {
	MonorepoGitDir string
	MonorepoHome   string
	MountHome      string // Reserved for single mount point
	Scm            string
	UserHandle     string
}

func LoadConfig() (*NdConfig, error) {
	if false {
		err := godotenv.Load()
		if err != nil {
			return nil, seederr.WrapErrorf("%w", err)
		}
	}
	monorepoHome := flagMonorepoHome.Get()
	if strings.HasPrefix(monorepoHome, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		monorepoHome = filepath.Join(homeDir, monorepoHome[2:])
	}
	monorepoGitDir := flagMonorepoGitDir.Get()
	if strings.HasPrefix(monorepoGitDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		monorepoGitDir = filepath.Join(homeDir, monorepoGitDir[2:])
	}
	userHandle := flagUserHandle.Get()
	ndConfig := &NdConfig{
		MountHome:      "",
		MonorepoHome:   monorepoHome,
		MonorepoGitDir: monorepoGitDir,
		Scm:            "",
		UserHandle:     userHandle,
	}
	ndConfig.Scm = flagScm.Get()
	if len(ndConfig.MonorepoHome) == 0 {
		return nil, seederr.WrapErrorf("ND_MONOREPO_HOME is not set")
	}
	if len(ndConfig.MonorepoGitDir) == 0 {
		return nil, seederr.WrapErrorf("ND_MONOREPO_GIT_DIR is not set")
	}
	if len(ndConfig.UserHandle) == 0 {
		return nil, seederr.WrapErrorf("ND_USER_HANDLE is not set")
	}
	return ndConfig, nil
}
