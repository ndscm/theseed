package common

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagDry = seedflag.DefineBool("dry", false, "make no external changes")
var flagScm = seedflag.DefineString("scm", "git", "the scm backend")
var flagShellEval = seedflag.DefineBool("shell-eval", false, "only output shell command")

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
	configHome, err := os.UserConfigDir()
	if err != nil {
		return nil, WrapTrace(err)
	}
	configPath := filepath.Join(configHome, "ndscm", "ndscm.env")
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		if flagDry.Get() {
			return nil, WrapTrace(fmt.Errorf("ndscm.env config is not found, remove --dry to initialize"))
		}
		err = os.MkdirAll(filepath.Dir(configPath), 0755)
		if err != nil {
			return nil, WrapTrace(err)
		}
		err = os.WriteFile(configPath, []byte{}, 0644)
		if err != nil {
			return nil, WrapTrace(err)
		}
	}
	err = godotenv.Load(configPath)
	if err != nil {
		return nil, WrapTrace(err)
	}
	ndConfig := &NdConfig{
		Dry:            false,
		MountHome:      "",
		MonorepoHome:   os.Getenv("ND_MONOREPO_HOME"),
		MonorepoGitDir: os.Getenv("ND_MONOREPO_GIT_DIR"),
		Scm:            "",
		ShellEval:      false,
		UserHandle:     os.Getenv("ND_USER_HANDLE"),
	}
	ndConfig.Dry = flagDry.Get()
	ndConfig.Scm = flagScm.Get()
	ndConfig.ShellEval = flagShellEval.Get()
	if len(ndConfig.MonorepoHome) == 0 {
		return nil, WrapTrace(fmt.Errorf("ND_MONOREPO_HOME is not set, please set it in %v", configPath))
	}
	if len(ndConfig.MonorepoGitDir) == 0 {
		return nil, WrapTrace(fmt.Errorf("ND_MONOREPO_GIT_DIR is not set, please set it in %v", configPath))
	}
	if len(ndConfig.UserHandle) == 0 {
		return nil, WrapTrace(fmt.Errorf("ND_USER_HANDLE is not set, please set it in %v", configPath))
	}
	return ndConfig, nil
}
