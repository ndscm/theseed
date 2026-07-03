package git

import (
	"encoding/json"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func getWipStatusPath(worktreePath string) (string, error) {
	gitArgs := []string{}
	if worktreePath != "" {
		gitArgs = append(gitArgs, "-C", worktreePath)
	}
	gitDirOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-parse", "--absolute-git-dir")...)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	gitDir := strings.TrimSpace(string(gitDirOutput))
	return filepath.Join(gitDir, "wip.ndscm.json"), nil
}

func LoadWipStatus(worktreePath string) (*scm.WipStatus, error) {
	statusPath, err := getWipStatusPath(worktreePath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	statusBytes, err := os.ReadFile(statusPath)
	if err != nil {
		if os.IsNotExist(err) {
			return nil, nil
		}
		return nil, seederr.WrapErrorf("failed to read wip status %v: %w", statusPath, err)
	}
	status := &scm.WipStatus{}
	err = json.Unmarshal(statusBytes, status)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to parse wip status %v: %w", statusPath, err)
	}
	return status, nil
}

func SaveWipStatus(worktreePath string, status *scm.WipStatus) error {
	statusPath, err := getWipStatusPath(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	statusBytes, err := json.MarshalIndent(status, "", "  ")
	if err != nil {
		return seederr.WrapErrorf("failed to marshal wip status: %w", err)
	}
	err = os.WriteFile(statusPath, statusBytes, 0644)
	if err != nil {
		return seederr.WrapErrorf("failed to write wip status %v: %w", statusPath, err)
	}
	return nil
}

func RemoveWipStatus(worktreePath string, force bool) error {
	statusPath, err := getWipStatusPath(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.Remove(statusPath)
	if err != nil {
		if os.IsNotExist(err) && force {
			return nil
		}
		return seederr.WrapErrorf("failed to remove wip status %v: %w", statusPath, err)
	}
	return nil
}
