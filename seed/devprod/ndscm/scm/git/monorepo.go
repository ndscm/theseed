package git

import (
	"path/filepath"
)

func guessMonorepoGitDir(monorepoHome string) string {
	repoIdentifier := filepath.Base(monorepoHome)
	monorepoGitDir := filepath.Join(monorepoHome, repoIdentifier+".git")
	return monorepoGitDir
}
