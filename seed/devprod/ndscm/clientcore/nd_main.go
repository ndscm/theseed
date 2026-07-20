package clientcore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdMainOptions struct {
	Remove bool

	Orphan string

	Area  string
	Start string
}

func getAreaWorktree(
	monorepoHome string, area string,
) (string, string, bool) {
	worktreeName := area + "/main"
	worktreePath := filepath.Join(monorepoHome, worktreeName)

	exists := false
	worktreeStat, err := os.Stat(worktreePath)
	if err == nil && worktreeStat.IsDir() {
		exists = true
	}

	return worktreeName, worktreePath, exists
}

func createAreaWorktree(
	scmProvider scm.Provider,
	monorepoHome string, area string, orphan string, startPoint string,
) (string, error) {
	worktreeName, worktreePath, exists := getAreaWorktree(monorepoHome, area)
	if exists {
		return "", seederr.WrapErrorf("worktree path already exists. path=%v", worktreePath)
	}
	branchName := worktreeName
	remote := "origin"
	remoteBranchName := branchName
	remoteTracking := remote + "/" + remoteBranchName
	err := scmProvider.FetchAll()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	remoteBranches, err := scmProvider.ListRemoteBranches(remote)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if slices.Contains(remoteBranches, remoteTracking) {
		if startPoint != "" {
			return "", seederr.WrapErrorf("cannot specify a start point when remote branch %v already exists", remoteTracking)
		}
		if orphan != "" {
			seedlog.Warnf("Ignoring --orphan message: remote branch %v already exists", remoteTracking)
		}
		err = scmProvider.CreateBranch(branchName, remoteTracking, remoteTracking)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create branch %v: %v", branchName, err)
		}
	} else {
		if startPoint == "" && orphan != "" {
			err = scmProvider.CreateOrphanBranch(branchName, orphan)
			if err != nil {
				return "", seederr.WrapErrorf("failed to create orphan branch %v: %v", branchName, err)
			}
		} else {
			if orphan != "" {
				seedlog.Warnf("Ignoring --orphan message: start point %v was specified", startPoint)
			}
			if startPoint == "" {
				startPoint = "origin/main"
			}
			err = scmProvider.CreateBranch(branchName, startPoint, "")
			if err != nil {
				return "", seederr.WrapErrorf("failed to create branch %v: %v", branchName, err)
			}
		}
		err = scmProvider.PushBranch(branchName, remote, remoteBranchName)
		if err != nil {
			return "", seederr.WrapErrorf("failed to push branch %v to %v: %v", branchName, remote, err)
		}
		err = scmProvider.SetBranchTracking(branchName, remoteTracking)
		if err != nil {
			return "", seederr.WrapErrorf("failed to set tracking for branch %v: %v", branchName, err)
		}
	}
	newWorktreePath, err := scmProvider.CreateWorktree(monorepoHome, branchName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create worktree for branch %v: %v", branchName, err)
	}
	if newWorktreePath != worktreePath {
		return "", seederr.WrapErrorf("unexpected new worktree path: %v (expected: %v)", newWorktreePath, worktreePath)
	}
	return newWorktreePath, nil
}

func removeAreaWorktree(
	scmProvider scm.Provider,
	monorepoHome string, area string,
) (string, error) {
	worktreeName, worktreePath, _ := getAreaWorktree(monorepoHome, area)
	branchName := worktreeName
	remoteTracking := "origin/" + branchName
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
	tracking, err := scmProvider.GetBranchTracking(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if tracking != remoteTracking {
		seedlog.Warnf("Area branch %v is tracking %v instead of %v, please cleanup before removing", branchName, tracking, remoteTracking)
		return "", seederr.WrapErrorf("area branch %v is not tracking %v", branchName, remoteTracking)
	}
	areaCommits, err := scmProvider.ListCommitIds(tracking, branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if len(areaCommits) > 0 {
		seedlog.Warnf("Area branch %v is not fully pushed to its tracking upstream %v, please drop the commits: %v", branchName, tracking, areaCommits)
		return "", seederr.WrapErrorf("area branch %v contains unpushed changes", branchName)
	}
	newCwd := ""
	if needChdir {
		mainWorktreePath := filepath.Join(monorepoHome, "main")
		err = os.Chdir(mainWorktreePath)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		newCwd = mainWorktreePath
	}
	err = scmProvider.RemoveWorktree(monorepoHome, worktreeName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return newCwd, nil
}

var areaRegex = regexp.MustCompile(`^[a-z0-9_-]+$`)

func NdMain(scmProvider scm.Provider, options NdMainOptions) error {
	if !seedshell.ShellEval() {
		seedlog.Warnf("It's recommended to run nd-main with --shell-eval")
	}
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.QuickVerifyMonorepo()
	if err != nil {
		return seederr.Wrap(err)
	}
	area := options.Area
	startPoint := options.Start
	worktreePath := ""
	if area == "" || area == "main" {
		if options.Remove {
			return seederr.WrapErrorf("cannot remove the main worktree")
		}
		if startPoint != "" {
			return seederr.WrapErrorf("cannot specify a start point for main area")
		}
		if options.Orphan != "" {
			seedlog.Warnf("Ignoring --orphan message: cannot create an orphan branch for the main area")
		}
		worktreePath = filepath.Join(monorepoHome, "main")
	} else {
		if !areaRegex.MatchString(area) {
			return seederr.WrapErrorf("only letters, digits, - and _ are allowed for area")
		}
		_, areaWorktreePath, exists := getAreaWorktree(monorepoHome, area)
		if options.Remove {
			if startPoint != "" {
				return seederr.WrapErrorf("cannot specify a start point with --remove")
			}
			if !exists {
				return seederr.WrapErrorf("area worktree %v does not exist", areaWorktreePath)
			}
			if options.Orphan != "" {
				seedlog.Warnf("Ignoring --orphan message: --remove was specified")
			}
			newCwd, err := removeAreaWorktree(scmProvider, monorepoHome, area)
			if err != nil {
				return seederr.Wrap(err)
			}
			if newCwd != "" {
				worktreePath = newCwd
			}
		} else {
			if exists {
				if startPoint != "" {
					return seederr.WrapErrorf("cannot specify a start point when area worktree already exists")
				}
				if options.Orphan != "" {
					seedlog.Warnf("Ignoring --orphan message: area worktree %v already exists", areaWorktreePath)
				}
			} else {
				areaWorktreePath, err = createAreaWorktree(scmProvider, monorepoHome, area, options.Orphan, startPoint)
				if err != nil {
					return seederr.Wrap(err)
				}
			}
			worktreePath = areaWorktreePath
		}
	}
	if worktreePath != "" {
		shellEval := fmt.Sprintf("\ncd \"%v\"\n", worktreePath)
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
