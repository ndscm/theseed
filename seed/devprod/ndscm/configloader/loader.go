package configloader

import (
	"os"
	"path/filepath"
	"regexp"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/tson/go/tson"
)

func FindDirConfigs(worktreePath string, sourcePaths []string) []string {
	result := []string{}
	for _, sourcePath := range sourcePaths {
		if filepath.Base(sourcePath) != "DIR.ndscm.ts" {
			continue
		}
		dirConfigPath := filepath.Join(worktreePath, sourcePath)
		_, err := os.Stat(dirConfigPath)
		if os.IsNotExist(err) {
			// The config was deleted (e.g. removed in the most recent commit)
			// but may still appear in the scan paths so other rules can react
			// to the deletion. There is nothing to load here.
			continue
		}
		result = append(result, sourcePath)
	}
	return result
}

func LoadDirConfig(worktreePath string, dirConfigPath string) (*DirConfig, error) {
	realPath := filepath.Join(worktreePath, dirConfigPath)
	seedlog.Debugf("Loading dir config: %s", realPath)
	raw, err := os.ReadFile(realPath)
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

func LoadDirConfigs(worktreePath string, dirConfigPaths []string) (map[string]*DirConfig, error) {
	result := map[string]*DirConfig{}
	for _, dirConfigPath := range dirConfigPaths {
		if filepath.Base(dirConfigPath) != "DIR.ndscm.ts" {
			return nil, seederr.WrapErrorf("invalid dir config path: %v", dirConfigPath)
		}
		dirConfig, err := LoadDirConfig(worktreePath, dirConfigPath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		result[dirConfigPath] = dirConfig
	}
	return result, nil
}

var upstreamNameRegex = regexp.MustCompile(`^[0-9A-Za-z]([0-9A-Za-z-]*[0-9A-Za-z])?$`)

func LoadRepoConfig(worktreePath string) (*RepoConfig, error) {
	repoConfigPath := filepath.Join(worktreePath, "REPO.ndscm.ts")
	seedlog.Debugf("Loading repo config: %s", repoConfigPath)
	raw, err := os.ReadFile(repoConfigPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
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

func FindUpstreamConfigs(worktreePath string, sourcePaths []string) []string {
	result := []string{}
	for _, sourcePath := range sourcePaths {
		if filepath.Base(sourcePath) != "UPSTREAM.ndscm.ts" {
			continue
		}
		upstreamConfigPath := filepath.Join(worktreePath, sourcePath)
		_, err := os.Stat(upstreamConfigPath)
		if os.IsNotExist(err) {
			// The config was deleted (e.g. removed in the most recent commit)
			// but may still appear in the scan paths so other rules can react
			// to the deletion. There is nothing to load here.
			continue
		}
		result = append(result, sourcePath)
	}
	return result
}

func LoadUpstreamConfig(worktreePath string, upstreamConfigPath string) (map[string]UpstreamConfig, error) {
	realPath := filepath.Join(worktreePath, upstreamConfigPath)
	seedlog.Debugf("Loading upstream config: %s", realPath)
	raw, err := os.ReadFile(realPath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	upstreamConfig := map[string]UpstreamConfig{}
	err = tson.Unmarshal(raw, &upstreamConfig)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	for upstreamName, config := range upstreamConfig {
		if upstreamName == "" {
			return nil, seederr.WrapErrorf("upstream name cannot be empty")
		}
		if config.Converge != "melt" {
			return nil, seederr.WrapErrorf("dir upstream converge must be `melt`: %v", upstreamName)
		}
		if config.Scm != "git" {
			return nil, seederr.WrapErrorf("dir upstream scm must be `git`: %v", upstreamName)
		}
	}
	return upstreamConfig, nil
}

func LoadUpstreamConfigs(worktreePath string, upstreamConfigPaths []string) (map[string]map[string]UpstreamConfig, error) {
	result := map[string]map[string]UpstreamConfig{}
	for _, upstreamConfigPath := range upstreamConfigPaths {
		if filepath.Base(upstreamConfigPath) != "UPSTREAM.ndscm.ts" {
			return nil, seederr.WrapErrorf("invalid upstream config path: %v", upstreamConfigPath)
		}
		upstreamConfig, err := LoadUpstreamConfig(worktreePath, upstreamConfigPath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		result[upstreamConfigPath] = upstreamConfig
	}
	return result, nil
}
