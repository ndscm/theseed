package clientcore

import (
	"fmt"
	"os"
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
	_, worktreePath, _ := scmProvider.GetDevWorktree(monorepoHome, currentUserHandle, focus, scm.CanonicalBranch())
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
		newCwd, err := scmProvider.RemoveDevWorktree(
			monorepoHome, currentUserHandle, focus, scm.CanonicalBranch(),
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
		newWorktreePath, err := scmProvider.CreateDevWorktree(
			monorepoHome, currentUserHandle, focus, tracking, scm.CanonicalBranch(),
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
