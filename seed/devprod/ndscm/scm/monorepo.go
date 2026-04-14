package scm

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagMonorepoHome = seedflag.DefineString("monorepo_home", "", "the monorepo home directory")
var flagMonorepoGitDir = seedflag.DefineString("monorepo_git_dir", "", "the monorepo git directory")

func MonorepoHome() (string, error) {
	monorepoHome := flagMonorepoHome.Get()
	if strings.HasPrefix(monorepoHome, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", seederr.Wrap(err)
		}
		monorepoHome = filepath.Join(homeDir, monorepoHome[2:])
	}
	if monorepoHome == "" {
		return "", seederr.WrapErrorf("monorepo home is not set")
	}
	return monorepoHome, nil
}

func MonorepoGitDir() (string, error) {
	monorepoGitDir := flagMonorepoGitDir.Get()
	if strings.HasPrefix(monorepoGitDir, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return "", seederr.Wrap(err)
		}
		monorepoGitDir = filepath.Join(homeDir, monorepoGitDir[2:])
	}
	if monorepoGitDir == "" {
		return "", seederr.WrapErrorf("monorepo git dir is not set")
	}
	return monorepoGitDir, nil
}
