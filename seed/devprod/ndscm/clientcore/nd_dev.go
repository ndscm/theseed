package clientcore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdDevOptions struct {
	Remove bool

	Track string
	Focus string
}

func getDevWorktree(
	monorepoHome string, ownerHandle string, focus string,
) (string, string, bool) {
	worktreeName := ""
	if focus == "" {
		focus = "main"
	}
	worktreeName = ownerHandle + "/dev/" + focus
	worktreePath := filepath.Join(monorepoHome, worktreeName)

	exists := false
	worktreeStat, err := os.Stat(worktreePath)
	if err == nil && worktreeStat.IsDir() {
		exists = true
	}

	return worktreeName, worktreePath, exists
}

func createDevWorktree(
	scmProvider scm.Provider,
	monorepoHome string, ownerHandle string, focus string, tracking string,
) (string, error) {
	worktreeName, worktreePath, exists := getDevWorktree(monorepoHome, ownerHandle, focus)
	if exists {
		return "", seederr.WrapErrorf("worktree path already exists. path=%v", worktreePath)
	}
	branchName := worktreeName
	baseBranchName, err := scm.GetBaseBranchName(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = scmProvider.CreateBranch(baseBranchName, tracking, tracking)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create base branch %v: %v", baseBranchName, err)
	}
	err = scmProvider.CreateBranch(branchName, baseBranchName, baseBranchName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create worktree branch %v: %v", branchName, err)
	}
	newWorktreePath, err := scmProvider.CreateWorktree(monorepoHome, branchName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create branch worktree %v: %v", branchName, err)
	}
	return newWorktreePath, nil
}

func removeDevWorktree(
	scmProvider scm.Provider,
	monorepoHome string, ownerHandle string, focus string,
) (string, error) {
	worktreeName, worktreePath, _ := getDevWorktree(monorepoHome, ownerHandle, focus)
	branchName := worktreeName
	baseBranchName, err := scm.GetBaseBranchName(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	_, currentWorktreePath, err := scmProvider.GetCurrentWorktree(monorepoHome)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	needChdir := currentWorktreePath == worktreePath
	dirtyFiles, err := scmProvider.ListDirtyFiles(worktreePath)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if len(dirtyFiles) > 0 {
		return "", seederr.WrapErrorf("workspace is dirty:\n%v", dirtyFiles)
	}
	devTracking, err := scmProvider.GetBranchTracking(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if devTracking != baseBranchName {
		seedlog.Warnf("Dev branch %v is tracking %v instead of its base branch, please cleanup dev worktree (with nd sync) before removing", branchName, devTracking)
		return "", seederr.WrapErrorf("dev branch %v is not tracking its base branch", branchName)
	}
	baseTracking, err := scmProvider.GetBranchTracking(baseBranchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	baseCommits, err := scmProvider.ListCommitIds(baseTracking, baseBranchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if len(baseCommits) > 0 {
		seedlog.Warnf("Base branch %v is not fully merged to its tracking upstream %v, please drop the commits: %v", baseBranchName, baseTracking, baseCommits)
		return "", seederr.WrapErrorf("base branch %v contains changes", baseBranchName)
	}
	devCommits, err := scmProvider.ListCommitIds(baseBranchName, branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if len(devCommits) > 0 {
		seedlog.Warnf("Dev branch %v is not fully merged to its base branch %v, please drop the commits: %v", branchName, baseBranchName, devCommits)
		return "", seederr.WrapErrorf("dev branch %v contains changes", branchName)
	}
	newCwd := ""
	if needChdir {
		err = os.Chdir(monorepoHome)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		newCwd = monorepoHome
	}
	err = scmProvider.RemoveWorktree(monorepoHome, worktreeName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(baseBranchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return newCwd, nil
}

var focusRegex = regexp.MustCompile(`^[a-zA-Z0-9_-]*$`)

func NdDev(scmProvider scm.Provider, options NdDevOptions) error {
	if !seedshell.ShellEval() {
		seedlog.Warnf("It's recommended to run nd-dev with --shell-eval")
	}
	currentUserHandle, err := user.CurrentUserHandle()
	if err != nil {
		return seederr.Wrap(err)
	}
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.QuickVerifyMonorepo()
	if err != nil {
		return seederr.Wrap(err)
	}
	track := options.Track
	focus := options.Focus
	if !focusRegex.MatchString(focus) {
		return seederr.WrapErrorf("only letters, digits, - and _ are allowed for focus")
	}
	_, worktreePath, _ := getDevWorktree(monorepoHome, currentUserHandle, focus)
	worktreeStat, err := os.Stat(worktreePath)
	if err != nil && !os.IsNotExist(err) {
		return seederr.WrapErrorf("failed to stat worktree %v: %v", worktreePath, err)
	}
	if err == nil && !worktreeStat.IsDir() {
		return seederr.WrapErrorf("worktree %v exists and is not a dir", worktreePath)
	}
	if options.Remove {
		if track != "" {
			return seederr.WrapErrorf("cannot specify --track with --remove")
		}
		if os.IsNotExist(err) {
			return seederr.WrapErrorf("dev worktree %v does not exist", worktreePath)
		}
		newCwd, err := removeDevWorktree(
			scmProvider, monorepoHome, currentUserHandle, focus,
		)
		if err != nil {
			return seederr.Wrap(err)
		}
		if newCwd != "" {
			shellEval := fmt.Sprintf("\ncd \"%v\"\n", newCwd)
			if seedshell.Dry() {
				seedlog.Infof("Shell eval: %v", shellEval)
			}
			if seedshell.ShellEval() {
				if !seedshell.Dry() {
					fmt.Printf("%v", shellEval)
				}
			}
		}
		return nil
	}
	if os.IsNotExist(err) {
		tracking := track
		if tracking == "" {
			tracking = "origin/main"
		}
		newWorktreePath, err := createDevWorktree(
			scmProvider, monorepoHome, currentUserHandle, focus, tracking,
		)
		if err != nil {
			return seederr.Wrap(err)
		}
		if newWorktreePath != worktreePath {
			return seederr.WrapErrorf("unexpected new worktree path: %v (expected: %v)", newWorktreePath, worktreePath)
		}
	} else {
		if track != "" {
			return seederr.WrapErrorf("cannot specify --track when dev worktree already exists")
		}
	}
	shellEval := fmt.Sprintf("\ncd \"%v\"\n", worktreePath)
	if seedshell.Dry() {
		seedlog.Infof("Shell eval: %v", shellEval)
	}
	if seedshell.ShellEval() {
		if !seedshell.Dry() {
			fmt.Printf("%v", shellEval)
		}
	}
	return nil
}
