package git

import (
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func CreateBranch(gitDir string, branchName string, startPoint string, tracking string) error {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "branch", branchName, startPoint)...)
	if err != nil {
		return seederr.WrapErrorf("failed to create branch %v: %w", branchName, err)
	}
	if tracking != "" {
		err = seedshell.ImpureRun("git", append(gitArgs, "branch", "--set-upstream-to="+tracking, branchName)...)
		if err != nil {
			return seederr.WrapErrorf("failed to set tracking for branch %v: %w", branchName, err)
		}
	}
	return nil
}

func DeleteBranch(gitDir string, branchName string) error {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "branch", "--delete", "--force", branchName)...)
	if err != nil {
		return seederr.WrapErrorf("failed to delete branch %v: %w", branchName, err)
	}
	return nil
}

func DeleteMergedBranch(gitDir string, branchName string) error {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "branch", "--delete", branchName)...)
	if err != nil {
		return seederr.WrapErrorf("failed to delete branch %v: %w", branchName, err)
	}
	return nil
}

func GetBranchTracking(gitDir string, branchName string) (string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	trackingBranchOutput, err := seedshell.PureOutput("git", append(gitArgs, "rev-parse", "--abbrev-ref", "--symbolic-full-name", branchName+"@{upstream}")...)
	if err != nil {
		return "", seederr.WrapErrorf("tracking upstream is missing for %v: %w", branchName, err)
	}
	trackingBranch := strings.TrimSpace(string(trackingBranchOutput))
	return trackingBranch, nil
}

func SetBranchTracking(gitDir string, branchName string, tracking string) error {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	err := seedshell.ImpureRun("git", append(gitArgs, "branch", "--set-upstream-to="+tracking, branchName)...)
	if err != nil {
		return seederr.WrapErrorf("failed to set tracking for branch %v: %w", branchName, err)
	}
	return nil
}

func GetMergeBaseHash(gitDir string, base string, target string) (string, error) {
	gitArgs := []string{}
	if gitDir != "" {
		gitArgs = append(gitArgs, "--git-dir", gitDir)
	}
	mergeBaseOutput, err := seedshell.PureOutput("git", append(gitArgs, "merge-base", base, target)...)
	if err != nil {
		return "", seederr.WrapErrorf("failed to get merge base for %v and %v: %w", base, target, err)
	}
	mergeBase := strings.TrimSpace(string(mergeBaseOutput))
	return mergeBase, nil
}
