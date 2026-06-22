package clientcore

import (
	"fmt"
	"os"
	"regexp"

	"github.com/ndscm/theseed/seed/devprod/ndscm/configloader"
	"github.com/ndscm/theseed/seed/devprod/ndscm/runner"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/devprod/scalpel/go/scalpel"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdMeltOptions struct {
	Remove bool
	Track  string

	Lock string

	Upstream string
	Commit   string
}

func NdMelt(scmProvider scm.Provider, options NdMeltOptions) error {
	if !seedshell.ShellEval() {
		seedlog.Warnf("It's recommended to run nd-melt with --shell-eval")
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
	upstreamName := options.Upstream
	if upstreamName == "" {
		return seederr.WrapErrorf("upstream name is required")
	}
	_, devWorktreePath, err := scmProvider.GetCurrentWorktree(monorepoHome)
	if err != nil {
		return seederr.Wrap(err)
	}
	repoConfig, err := configloader.LoadRepoConfig(devWorktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	upstreamConfig, ok := repoConfig.Upstream[upstreamName]
	if !ok {
		return seederr.WrapErrorf("upstream %v not found in repo config", upstreamName)
	}
	if upstreamConfig.Converge != "melt" {
		return seederr.WrapErrorf("upstream %v is not a melt upstream", upstreamName)
	}
	if upstreamConfig.Local {
		return seederr.WrapErrorf("upstream %v is a local upstream, cannot melt", upstreamName)
	}
	if upstreamConfig.Repo == "" {
		return seederr.WrapErrorf("upstream %v does not have a remote specified", upstreamName)
	}
	if upstreamConfig.Tracking == "" {
		return seederr.WrapErrorf("upstream %v does not have a tracking branch specified", upstreamName)
	}
	err = scmProvider.FetchBranch(upstreamConfig.Repo, upstreamConfig.Tracking, "upstream/"+upstreamName)
	if err != nil {
		return seederr.Wrap(err)
	}
	_, worktreePath, _ := scmProvider.GetMeltWorktree(
		monorepoHome, currentUserHandle, upstreamName,
	)
	worktreeStat, err := os.Stat(worktreePath)
	if err != nil && !os.IsNotExist(err) {
		return seederr.WrapErrorf("failed to stat worktree %v: %v", worktreePath, err)
	}
	if err == nil && !worktreeStat.IsDir() {
		return seederr.WrapErrorf("worktree %v exists and is not a dir", worktreePath)
	}
	worktreeExists := err == nil

	if options.Remove {
		if !worktreeExists {
			return seederr.WrapErrorf("melt worktree %v does not exist", worktreePath)
		}
		newCwd, err := scmProvider.RemoveMeltWorktree(
			monorepoHome, currentUserHandle, upstreamName,
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

	if !worktreeExists {
		tracking := track
		if tracking == "" {
			tracking = "origin/main"
		}
		trackingTipPoint := tracking
		upstreamTipPoint := options.Commit
		if upstreamTipPoint == "" {
			upstreamTipPoint = "upstream/" + upstreamName
		}
		trackingForkPoint, upstreamForkPoint, err := scmProvider.SearchForkPoint(trackingTipPoint, upstreamTipPoint)
		if err != nil {
			return seederr.Wrap(err)
		}
		seedlog.Infof("Found tracking fork point: %v", trackingForkPoint)
		seedlog.Infof("Found upstream fork point: %v", upstreamForkPoint)
		newWorktreePath, err := scmProvider.CreateMeltWorktree(
			monorepoHome, currentUserHandle,
			upstreamName, upstreamForkPoint, tracking, trackingForkPoint,
		)
		if err != nil {
			return seederr.Wrap(err)
		}
		if newWorktreePath != worktreePath {
			return seederr.WrapErrorf("unexpected new worktree path: %v (expected: %v)", newWorktreePath, worktreePath)
		}

		dropLocks := [][]*regexp.Regexp{}
		switch options.Lock {
		case "drop":
			scmFilePaths, err := scmProvider.ListFiles(worktreePath)
			if err != nil {
				return seederr.Wrap(err)
			}
			repoAnalysis, err := runner.AnalyseRepo(worktreePath, []string{"lock"}, scmFilePaths, runner.BuildSystems())
			if err != nil {
				return seederr.Wrap(err)
			}
			lockPhase := repoAnalysis.Phases["lock"]
			for _, watcher := range lockPhase.Watchers {
				lockGroup := []*regexp.Regexp{}
				for _, lockFile := range watcher.Targets {
					lockGroup = append(lockGroup, regexp.MustCompile(`^`+regexp.QuoteMeta(lockFile)+`$`))
				}
				dropLocks = append(dropLocks, lockGroup)
			}
		case "keep":
			// pass
		default:
			return seederr.WrapErrorf("unknown lock strategy %q, want one of: drop, keep", options.Lock)
		}

		upstreamCommitIds, err := scmProvider.ListCommitIds(upstreamForkPoint, upstreamTipPoint)
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, commitId := range upstreamCommitIds {
			commitPatch, err := scmProvider.GetCommitFormatPatch(commitId)
			if err != nil {
				return seederr.Wrap(err)
			}
			patch, err := scalpel.ParseFormatPatch(commitId, commitPatch)
			if err != nil {
				return seederr.Wrap(err)
			}
			for _, lockGroup := range dropLocks {
				patch.DropDiff(lockGroup, lockGroup)
			}
			newCommitPatch := patch.Render()
			err = scmProvider.ApplyFormatPatch(newWorktreePath, newCommitPatch)
			if err != nil {
				return seederr.Wrap(err)
			}
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
