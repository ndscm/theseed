package git

import (
	"bytes"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type GitFileStatus struct {
	Status string
	To     string
	From   string
}

func (s GitFileStatus) String() string {
	if s.From != "" {
		return s.Status + " " + s.To + " <- " + s.From
	}
	return s.Status + " " + s.To
}

func GetStatus(worktreePath string) ([]GitFileStatus, error) {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	statusOutput, err := seedshell.PureOutput("git", append(gitArgs, "status", "--porcelain", "-z", "--untracked-files=all")...)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	lines := bytes.Split(statusOutput, []byte("\000"))
	renaming := false
	result := []GitFileStatus{}
	for _, line := range lines {
		if len(line) == 0 {
			continue
		}
		if renaming {
			result[len(result)-1].From = string(line)
			renaming = false
			continue
		}
		if len(line) < 4 {
			return nil, seederr.WrapErrorf("git status line is malformed: %v", string(line))
		}
		status := string(line[:2])
		to := string(line[3:])
		result = append(result, GitFileStatus{
			Status: status,
			To:     to,
			From:   "",
		})
		if strings.HasPrefix(status, "R") || strings.HasPrefix(status, "C") {
			renaming = true
		}
	}
	return result, nil
}

func GetCurrentBranch(worktreePath string) (string, error) {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	currentBranchOutput, err := seedshell.PureOutput("git", append(gitArgs, "branch", "--show-current")...)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	currentBranch := strings.TrimSpace(string(currentBranchOutput))
	return currentBranch, nil
}

func GetCurrentOperation(worktreePath string) (string, error) {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	gitDirOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-parse", "--git-dir")...)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	gitDir := strings.TrimSpace(string(gitDirOutput))
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-merge")); err == nil {
		return "rebase", nil
	}
	if _, err := os.Stat(filepath.Join(gitDir, "rebase-apply")); err == nil {
		return "rebase", nil
	}
	if _, err := os.Stat(filepath.Join(gitDir, "MERGE_HEAD")); err == nil {
		return "merge", nil
	}
	if _, err := os.Stat(filepath.Join(gitDir, "CHERRY_PICK_HEAD")); err == nil {
		return "cherry-pick", nil
	}
	return "", nil
}
