package configloader

import (
	"os"
	"path/filepath"

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

func LoadDirConfigs(worktree string, scmFilePaths []string) (map[string]*DirConfig, error) {
	result := map[string]*DirConfig{}
	for _, scmFilePath := range scmFilePaths {
		if filepath.Base(scmFilePath) != "DIR.ndscm.ts" {
			continue
		}
		dirConfig, err := LoadDirConfig(filepath.Join(worktree, scmFilePath))
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		result[scmFilePath] = dirConfig
	}
	return result, nil
}

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
	return repoConfig, nil
}
