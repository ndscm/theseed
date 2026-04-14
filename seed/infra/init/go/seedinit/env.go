package seedinit

import (
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/infra/dotenv/go/dotenv"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func loadSystemEnv(envPath string) error {
	systemConfigHome := ""
	_, err := os.Stat("/etc")
	if err == nil {
		systemConfigHome = "/etc"
	}

	if systemConfigHome == "" {
		return seederr.WrapErrorf("system config home is not found")
	}

	absEnvPath := filepath.Join(systemConfigHome, envPath)
	_, err = os.Stat(absEnvPath)
	if err != nil {
		// Ignore if the file does not exist
		return nil
	}

	err = dotenv.Load(absEnvPath)
	if err != nil {
		return seederr.Wrap(err)
	}

	return nil
}

func loadUserEnv(envPath string) error {
	userConfigHome, err := os.UserConfigDir()
	if err != nil {
		return seederr.Wrap(err)
	}

	absEnvPath := filepath.Join(userConfigHome, envPath)
	_, err = os.Stat(absEnvPath)
	if err != nil {
		// Ignore if the file does not exist
		return nil
	}

	err = dotenv.Load(absEnvPath)
	if err != nil {
		return seederr.Wrap(err)
	}

	return nil
}
