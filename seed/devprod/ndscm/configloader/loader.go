package configloader

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/tson/go/tson"
)

func LoadDirConfig(dirConfigPath string) (*DirConfig, error) {
	seedlog.Debugf("Loading dir config: %s", dirConfigPath)
	raw, err := os.ReadFile(dirConfigPath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	dirConfig := &DirConfig{}
	err = tson.Unmarshal(raw, dirConfig)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return dirConfig, nil
}

func LoadDirConfigs(worktreePath string, scmFilePaths []string) (map[string]*DirConfig, error) {
	result := map[string]*DirConfig{}
	for _, scmFilePath := range scmFilePaths {
		if filepath.Base(scmFilePath) != "DIR.ndscm.ts" {
			continue
		}
		dirConfigPath := filepath.Join(worktreePath, scmFilePath)
		_, err := os.Stat(dirConfigPath)
		if os.IsNotExist(err) {
			// The config was deleted (e.g. removed in the most recent commit)
			// but may still appear in the scan paths so other rules can react
			// to the deletion. There is nothing to load here.
			continue
		}
		dirConfig, err := LoadDirConfig(dirConfigPath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		result[scmFilePath] = dirConfig
	}
	return result, nil
}

var upstreamNameRegex = regexp.MustCompile(`^[0-9A-Za-z]([0-9A-Za-z-]*[0-9A-Za-z])?$`)

func LoadRepoConfig(worktreePath string) (*RepoConfig, error) {
	repoConfigPath := filepath.Join(worktreePath, "REPO.ndscm.ts")
	seedlog.Debugf("Loading repo config: %s", repoConfigPath)
	raw, err := os.ReadFile(repoConfigPath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	repoConfig := &RepoConfig{}
	err = tson.Unmarshal(raw, repoConfig)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	hasRebase := false
	for upstreamName, upstreamConfig := range repoConfig.Upstream {
		if !upstreamNameRegex.MatchString(upstreamName) {
			return nil, seederr.WrapErrorf("invalid upstream name: %v", upstreamName)
		}
		if upstreamConfig.Converge == "rebase" {
			if hasRebase {
				seedlog.Warnf("Multiple rebase upstreams found. It's recommended to have only one rebase upstream")
			}
			if upstreamName != "default" {
				seedlog.Warnf("It's recommended to set the upstream name to `default` for the rebase upstream")
			}
			hasRebase = true
		}
	}
	return repoConfig, nil
}
