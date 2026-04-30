package git

import (
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

// SetConfig writes a config entry to the repo-local config (git config defaults to --local).
func SetConfig(gitDir string, configName string, value string) error {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "config", configName, value)...)
	if err != nil {
		return seederr.WrapErrorf("failed to set config %v: %w", configName, err)
	}
	return nil
}
