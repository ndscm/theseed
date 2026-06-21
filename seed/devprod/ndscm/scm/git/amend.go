package git

import (
	"regexp"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

var trailerKeyRegex = regexp.MustCompile(`^([a-z0-9]+)(-[a-z0-9]+)*$`)

func AmendAppendTrailer(worktreePath string, trailerKey string, text string) error {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	gitArgs = append(gitArgs, "commit", "--amend", "--no-edit", "--trailer", trailerKey+": "+text)
	err := seedshell.ImpureRun("git", gitArgs...)
	if err != nil {
		return seederr.WrapErrorf("failed to amend head commit with trailer %v: %w", trailerKey, err)
	}
	return nil
}
