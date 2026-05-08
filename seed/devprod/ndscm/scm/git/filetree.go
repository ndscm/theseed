package git

import (
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func ListCommitFiles(gitDir string, commit string) ([]string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	listOutput, err := seedshell.PureOutput("git", append(gitArgs, "diff-tree", "--no-commit-id", "--name-only", "-r", commit)...)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to get commit files for %v: %w", commit, err)
	}
	trimmed := strings.TrimSpace(string(listOutput))
	if trimmed == "" {
		return nil, nil
	}
	files := strings.Split(trimmed, "\n")
	return files, nil
}
