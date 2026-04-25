package scm

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagMonorepoHome = seedflag.DefineString("monorepo_home", "", "the monorepo home directory")

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
