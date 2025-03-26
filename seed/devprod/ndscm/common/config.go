package common

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"

	"github.com/joho/godotenv"
)

var flagDry = flag.Bool("dry", false, "make no external changes")
var flagScm = flag.String("scm", "git", "the scm backend")
var flagShellEval = flag.Bool("shell-eval", false, "only output shell command")

type NdConfig struct {
	Dry            bool
	MonorepoGitDir string
	MonorepoHome   string
	MountHome      string // Reserved for single mount point
	Scm            string
	ShellEval      bool
}

func LoadConfig() (*NdConfig, error) {
	configHome, err := os.UserConfigDir()
	if err != nil {
		return nil, WrapTrace(err)
	}
	configPath := filepath.Join(configHome, "ndscm", "ndscm.env")
	_, err = os.Stat(configPath)
	if os.IsNotExist(err) {
		if *flagDry {
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
	}
	ndConfig.Dry = *flagDry
	ndConfig.Scm = *flagScm
	ndConfig.ShellEval = *flagShellEval
	return ndConfig, nil
}
