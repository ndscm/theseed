package clientcore

import (
	"fmt"
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type ndDevFlags struct {
	remove *seedflag.BoolFlag
	track  *seedflag.StringFlag
}

func parseNdDevFlags(args []string) (ndDevFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-dev")
	cmdFlags := ndDevFlags{}
	cmdFlags.remove = cf.DefineBool("remove", false, "Remove the dev worktree")
	cmdFlags.track = cf.DefineString("track", "", "Remote branch to track (e.g. origin/main); only valid when creating a new dev worktree, must not be provided if the dev worktree already exists")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func NdDev(scmProvider scm.Provider, args []string) error {
	if !seedshell.ShellEval() {
		seedlog.Warnf("It's recommended to run nd-dev with --shell-eval")
	}
	cmdFlags, cmdArgs, err := parseNdDevFlags(args)
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
	track := cmdFlags.track.Get()
	focus := ""
	if len(cmdArgs) == 0 {
		// pass
	} else if len(cmdArgs) == 1 {
		focus = cmdArgs[0]
	} else {
		return seederr.WrapErrorf("nd-dev usage: nd dev [<focus-area>]")
	}
	worktreePath := scmProvider.GetDevWorktree(monorepoHome, focus)
	worktreeStat, err := os.Stat(worktreePath)
	if err != nil && !os.IsNotExist(err) {
		return seederr.WrapErrorf("failed to stat worktree %v: %v", worktreePath, err)
	}
	if err == nil && !worktreeStat.IsDir() {
		return seederr.WrapErrorf("worktree %v exists and is not a dir", worktreePath)
	}
	if cmdFlags.remove.Get() {
		if track != "" {
			return seederr.WrapErrorf("cannot specify --track with --remove")
		}
		if os.IsNotExist(err) {
			return seederr.WrapErrorf("dev worktree %v does not exist", worktreePath)
		}
		newCwd, err := scmProvider.RemoveDevWorktree(monorepoHome, focus)
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
		newWorktreePath, err := scmProvider.CreateDevWorktree(monorepoHome, focus, tracking)
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
