package git

import (
	"bytes"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func ListCommitFiles(gitDir string, commit string) ([]GitFileStatus, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	listOutput, err := seedshell.PureOutput("git", append(gitArgs, "diff-tree", "--no-commit-id", "-r", "-z", commit)...)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to get commit files for %v: %w", commit, err)
	}
	fields := bytes.Split(listOutput, []byte("\000"))
	result := []GitFileStatus{}
	for len(fields) > 0 {
		if len(fields[0]) == 0 {
			fields = fields[1:]
			continue
		}
		status := string(fields[0])
		if len(fields) < 2 {
			return nil, seederr.WrapErrorf("malformed diff-tree output for %v", commit)
		}
		to := string(fields[1])
		fields = fields[2:]
		from := ""
		if strings.HasPrefix(status, "R") || strings.HasPrefix(status, "C") {
			from = to
			if len(fields) < 1 {
				return nil, seederr.WrapErrorf("malformed diff-tree rename output for %v", commit)
			}
			to = string(fields[0])
			fields = fields[1:]
		}
		result = append(result, GitFileStatus{
			Status: status,
			To:     to,
			From:   from,
		})
	}
	return result, nil
}
