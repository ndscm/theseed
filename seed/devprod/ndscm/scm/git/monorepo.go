package git

import (
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
)

func guessMonorepoGitDir(repo *scm.WorkingRepo) string {
	if repo == nil || repo.MonorepoHome == "" {
		return ""
	}
	repoIdentifier := filepath.Base(repo.MonorepoHome)
	monorepoGitDir := filepath.Join(repo.MonorepoHome, repoIdentifier+".git")
	return monorepoGitDir
}
