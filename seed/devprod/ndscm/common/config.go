package common

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/joho/godotenv"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagDry = seedflag.DefineBool("dry", false, "make no external changes")
var flagScm = seedflag.DefineString("scm", "git", "the scm backend")
var flagShellEval = seedflag.DefineBool("shell-eval", false, "only output shell command")

var flagMonorepoHome = seedflag.DefineString("monorepo_home", "", "the monorepo home directory")
var flagMonorepoGitDir = seedflag.DefineString("monorepo_git_dir", "", "the monorepo git directory")
var flagUserHandle = seedflag.DefineString("user_handle", "", "the user handle")

type NdConfig struct {
	Dry            bool
	MonorepoGitDir string
	MonorepoHome   string
	MountHome      string // Reserved for single mount point
	Scm            string
	ShellEval      bool
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
		Dry:            false,
		MountHome:      "",
		MonorepoHome:   monorepoHome,
		MonorepoGitDir: monorepoGitDir,
		Scm:            "",
		ShellEval:      false,
		UserHandle:     userHandle,
	}
	ndConfig.Dry = flagDry.Get()
	ndConfig.Scm = flagScm.Get()
	ndConfig.ShellEval = flagShellEval.Get()
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
