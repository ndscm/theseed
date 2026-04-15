package git

import (
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func GetCommitHash(gitDir string, commit string) (string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	commitOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-parse", commit)...)
	if err != nil {
		return "", seederr.WrapErrorf("failed to get commit hash for %v: %w", commit, err)
	}
	commitHash := strings.TrimSpace(string(commitOutput))
	return commitHash, nil
}

func ListCommitHash(gitDir string, from string, to string) ([]string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	listOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-list", "--ancestry-path", from+".."+to)...)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to get commit hash for range %v..%v: %w", from, to, err)
	}
	trimmed := strings.TrimSpace(string(listOutput))
	if trimmed == "" {
		return nil, nil
	}
	commits := strings.Split(trimmed, "\n")
	return commits, nil
}

func ListMergeCommitHash(gitDir string, from string, to string) ([]string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	listOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-list", "--merges", "--ancestry-path", from+".."+to)...)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to get merge commit hash for range %v..%v: %w", from, to, err)
	}
	trimmed := strings.TrimSpace(string(listOutput))
	if trimmed == "" {
		return nil, nil
	}
	commits := strings.Split(trimmed, "\n")
	return commits, nil
}
