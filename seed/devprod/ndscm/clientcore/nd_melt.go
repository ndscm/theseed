package clientcore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"strings"

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

func findUpstreamConfig(
	worktreePath string, sourceFiles []string, upstreamName string,
) (string, *configloader.UpstreamConfig, error) {
	upstreamConfigPaths := configloader.FindUpstreamConfigs(worktreePath, sourceFiles)
	upstreamConfigs, err := configloader.LoadUpstreamConfigs(worktreePath, upstreamConfigPaths)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}

	repoConfig, err := configloader.LoadRepoConfig(worktreePath)
	if err != nil {
		return "", nil, seederr.Wrap(err)
	}
	if repoConfig != nil && repoConfig.Upstream != nil {
		upstreamConfigs["REPO.ndscm.ts"] = repoConfig.Upstream
	}

	foundConfigPath := ""
	foundConfig := (*configloader.UpstreamConfig)(nil)
	for configPath, upstreams := range upstreamConfigs {
		found, ok := upstreams[upstreamName]
		if !ok {
			continue
		}
		if foundConfig != nil {
			return "", nil, seederr.WrapErrorf(
				"upstream %v is defined in multiple configs: %v, %v",
				upstreamName, foundConfigPath, configPath,
			)
		}
		foundConfigPath = configPath
		foundConfig = &found
	}

	if foundConfigPath == "" {
		return "", nil, seederr.WrapErrorf("upstream %v not found in any upstream config", upstreamName)
	}
	return foundConfigPath, foundConfig, nil
}

type simpleRenamePattern struct {
	match *regexp.Regexp

	replace string
}

type meltModifiers struct {
	includePatterns []*regexp.Regexp

	excludePatterns []*regexp.Regexp

	renamePatterns []simpleRenamePattern
}

func (m *meltModifiers) apply(path string) string {
	if len(m.includePatterns) > 0 {
		included := false
		for _, re := range m.includePatterns {
			if re.MatchString(path) {
				included = true
				break
			}
		}
		if !included {
			return ""
		}
	}
	for _, re := range m.excludePatterns {
		if re.MatchString(path) {
			return ""
		}
	}
	for _, renamePattern := range m.renamePatterns {
		if renamePattern.match.MatchString(path) {
			path = renamePattern.match.ReplaceAllString(path, renamePattern.replace)
		}
	}
	return path
}

func compileMeltModifiers(upstreamConfig *configloader.UpstreamConfig) (*meltModifiers, error) {
	includePatterns := []*regexp.Regexp{}
	for _, pattern := range upstreamConfig.Include {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, seederr.WrapErrorf("invalid include pattern %q: %v", pattern, err)
		}
		includePatterns = append(includePatterns, re)
	}
	excludePatterns := []*regexp.Regexp{}
	for _, pattern := range upstreamConfig.Exclude {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, seederr.WrapErrorf("invalid exclude pattern %q: %v", pattern, err)
		}
		excludePatterns = append(excludePatterns, re)
	}
	renamePatterns := []simpleRenamePattern{}
	for _, rule := range upstreamConfig.Rename {
		re, err := regexp.Compile(rule.Match)
		if err != nil {
			return nil, seederr.WrapErrorf("invalid rename match pattern %q: %v", rule.Match, err)
		}
		renamePatterns = append(renamePatterns, simpleRenamePattern{match: re, replace: rule.Replace})
	}

	if len(includePatterns) == 0 && len(excludePatterns) == 0 && len(renamePatterns) == 0 {
		return nil, nil
	}

	return &meltModifiers{
		includePatterns: includePatterns,
		excludePatterns: excludePatterns,
		renamePatterns:  renamePatterns,
	}, nil
}

// constructMeltFromPoint builds a commit on top of upstreamForkPoint whose tree
// has the melt modifiers and upstreamDir prefix already applied, so the rewritten
// upstream patches apply cleanly on it. It anchors the commit on a scratch branch
// and returns that branch name for the caller to resolve and later tear down.
func constructMeltFromPoint(
	scmProvider scm.Provider,
	repo *scm.WorkingRepo, ownerHandle string,
	upstreamName string, upstreamForkPoint string, upstreamDir string, modifiers *meltModifiers,
) (string, error) {
	scratchBranch := ownerHandle + "/wip/melt-scratch-" + upstreamName

	// Clear any scratch branch/worktree left behind by a failed prior run.
	err := scmProvider.RemoveWorktree(repo, scratchBranch)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(repo, scratchBranch)
	if err != nil {
		return "", seederr.Wrap(err)
	}

	err = scmProvider.CreateBranch(scratchBranch, upstreamForkPoint, "")
	if err != nil {
		return "", seederr.WrapErrorf("failed to create melt fork point branch %v: %v", scratchBranch, err)
	}
	scratchWorktreePath, err := scmProvider.CreateWorktree(repo, scratchBranch)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create melt fork point worktree %v: %v", scratchBranch, err)
	}

	scmFilePaths, err := scmProvider.ListFiles(scratchWorktreePath)
	if err != nil {
		return "", seederr.Wrap(err)
	}

	changed := false
	for _, scmFilePath := range scmFilePaths {
		newPath := scmFilePath
		if modifiers != nil {
			newPath = modifiers.apply(scmFilePath)
		}
		if newPath == "" {
			err = os.Remove(filepath.Join(scratchWorktreePath, scmFilePath))
			if err != nil {
				return "", seederr.WrapErrorf("failed to drop %v: %v", scmFilePath, err)
			}
			changed = true
			continue
		}
		if upstreamDir != "." {
			newPath = upstreamDir + "/" + newPath
		}
		if newPath == scmFilePath {
			continue
		}
		newAbsPath := filepath.Join(scratchWorktreePath, newPath)
		err = os.MkdirAll(filepath.Dir(newAbsPath), 0755)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create directory for %v: %v", newPath, err)
		}
		err = os.Rename(filepath.Join(scratchWorktreePath, scmFilePath), newAbsPath)
		if err != nil {
			return "", seederr.WrapErrorf("failed to move %v to %v: %v", scmFilePath, newPath, err)
		}
		changed = true
	}

	result := ""
	if changed {
		err = scmProvider.CreateCommit(scratchWorktreePath, "melt: prepare")
		if err != nil {
			return "", seederr.Wrap(err)
		}
		result = scratchBranch
	}

	err = scmProvider.RemoveWorktree(repo, scratchBranch)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if !changed {
		err = scmProvider.DeleteBranch(repo, scratchBranch)
		if err != nil {
			return "", seederr.Wrap(err)
		}
	}
	return result, nil
}

func getMeltWorktree(
	monorepoHome string, ownerHandle string, upstreamName string,
) (string, string, bool) {
	if upstreamName == "" {
		upstreamName = "default"
	}
	worktreeName := ownerHandle + "/melt/" + upstreamName
	worktreePath := filepath.Join(monorepoHome, worktreeName)

	exists := false
	worktreeStat, err := os.Stat(worktreePath)
	if err == nil && worktreeStat.IsDir() {
		exists = true
	}

	return worktreeName, worktreePath, exists
}

func createMeltWorktree(
	scmProvider scm.Provider,
	repo *scm.WorkingRepo, ownerHandle string,
	upstreamName string, fromPoint string, tracking string, forkPoint string,
) (string, error) {
	if upstreamName == "" {
		return "", seederr.WrapErrorf("upstream name is required for melt worktree")
	}
	worktreeName, worktreePath, exists := getMeltWorktree(repo.MonorepoHome, ownerHandle, upstreamName)
	if exists {
		return "", seederr.WrapErrorf("worktree path already exists. path=%v", worktreePath)
	}
	branchName := worktreeName
	baseBranchName, err := scm.GetBaseBranchName(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	meltWorktreePath := ""
	if fromPoint == "" {
		// disjoint: no shared upstream fork point. Create the base branch at the
		// tracking initial commit (forkPoint). The worktree branch does not
		// exist yet, so CreateWorktree roots it as an orphan and all upstream
		// commits melt onto an empty tree; set its upstream to the base for nd
		// sync. No base update is needed.
		err = scmProvider.CreateBranch(baseBranchName, forkPoint, tracking)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create base branch %v: %v", baseBranchName, err)
		}
		newWorktreePath, err := scmProvider.CreateWorktree(repo, branchName)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create orphan worktree %v: %v", branchName, err)
		}
		err = scmProvider.SetBranchTracking(branchName, baseBranchName)
		if err != nil {
			return "", seederr.WrapErrorf("failed to set tracking for branch %v: %v", branchName, err)
		}
		meltWorktreePath = newWorktreePath
	} else {
		// Shared fork point: start the base and worktree branches at the
		// upstream-side fork (fromPoint), then move the base to the
		// tracking-side fork (forkPoint). The upstream commits after fromPoint
		// are melted on top, and the base's fromPoint->forkPoint reflog entry
		// tells nd sync where to rebase from.
		err = scmProvider.CreateBranch(baseBranchName, fromPoint, tracking)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create base branch %v: %v", baseBranchName, err)
		}
		err = scmProvider.CreateBranch(branchName, fromPoint, baseBranchName)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create worktree branch %v: %v", branchName, err)
		}
		newWorktreePath, err := scmProvider.CreateWorktree(repo, branchName)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create branch worktree %v: %v", branchName, err)
		}
		// Move base branch from fromPoint to forkPoint, creating a reflog
		// entry that nd sync (git pull --rebase) uses to rebase onto.
		err = scmProvider.UpdateBranch(baseBranchName, forkPoint)
		if err != nil {
			return "", seederr.WrapErrorf("failed to update base branch %v: %v", baseBranchName, err)
		}
		meltWorktreePath = newWorktreePath
	}
	return meltWorktreePath, nil
}

func removeMeltWorktree(
	scmProvider scm.Provider,
	repo *scm.WorkingRepo, ownerHandle string, upstreamName string,
) (string, error) {
	if upstreamName == "" {
		return "", seederr.WrapErrorf("upstream name is required for melt worktree")
	}
	worktreeName, worktreePath, _ := getMeltWorktree(repo.MonorepoHome, ownerHandle, upstreamName)
	branchName := worktreeName
	baseBranchName, err := scm.GetBaseBranchName(branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	_, currentWorktreePath, err := scmProvider.GetCurrentWorktree(repo)
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
	newCwd := ""
	if needChdir {
		err = os.Chdir(repo.MonorepoHome)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		newCwd = repo.MonorepoHome
	}
	err = scmProvider.RemoveWorktree(repo, worktreeName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(repo, branchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(repo, baseBranchName)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return newCwd, nil
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
	repo := &scm.WorkingRepo{MonorepoHome: monorepoHome}
	err = scmProvider.QuickVerifyMonorepo(repo)
	if err != nil {
		return seederr.Wrap(err)
	}
	track := options.Track
	upstreamName := options.Upstream
	if upstreamName == "" {
		return seederr.WrapErrorf("upstream name is required")
	}
	_, devWorktreePath, err := scmProvider.GetCurrentWorktree(repo)
	if err != nil {
		return seederr.Wrap(err)
	}

	scmFiles, err := scmProvider.ListFiles(devWorktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	upstreamConfigPath, upstreamConfig, err := findUpstreamConfig(devWorktreePath, scmFiles, upstreamName)
	if err != nil {
		return seederr.Wrap(err)
	}
	upstreamDir := filepath.Dir(upstreamConfigPath)

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
	_, worktreePath, _ := getMeltWorktree(
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
		newCwd, err := removeMeltWorktree(
			scmProvider, repo, currentUserHandle, upstreamName,
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
		if upstreamForkPoint != "" {
			seedlog.Infof("Found tracking fork point: %v", trackingForkPoint)
			seedlog.Infof("Found upstream fork point: %v", upstreamForkPoint)
		} else {
			seedlog.Infof("No fork point found. Melting from history root: %v", trackingForkPoint)
		}

		modifiers, err := compileMeltModifiers(upstreamConfig)
		if err != nil {
			return seederr.Wrap(err)
		}

		upstreamFromPoint := upstreamForkPoint
		scratchBranch := ""
		if (modifiers != nil || upstreamDir != ".") && upstreamForkPoint != "" {
			scratchBranch, err = constructMeltFromPoint(
				scmProvider, repo, currentUserHandle,
				upstreamName, upstreamForkPoint, upstreamDir, modifiers,
			)
			if err != nil {
				return seederr.Wrap(err)
			}
			if scratchBranch != "" {
				upstreamFromPoint, err = scmProvider.GetBranch(scratchBranch)
				if err != nil {
					return seederr.Wrap(err)
				}
			}
		}

		newWorktreePath, err := createMeltWorktree(
			scmProvider, repo, currentUserHandle,
			upstreamName, upstreamFromPoint, tracking, trackingForkPoint,
		)
		if err != nil {
			return seederr.Wrap(err)
		}
		if newWorktreePath != worktreePath {
			return seederr.WrapErrorf("unexpected new worktree path: %v (expected: %v)", newWorktreePath, worktreePath)
		}

		if scratchBranch != "" {
			err = scmProvider.DeleteBranch(repo, scratchBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
		}

		// dropLocks holds path patterns, grouped per lock watcher, for lock
		// files to strip from upstream diffs. "drop" removes them so they are
		// regenerated locally; "keep" lets upstream lockfiles flow through.
		dropLocks := [][]*regexp.Regexp{}
		switch options.Lock {
		case "drop":
			// Analyse the dev worktree, not the melt worktree. The melt worktree
			// holds the include-filtered upstream tree, which may omit the config
			// (e.g. DIR.ndscm.ts) that defines the lock rules; the dev worktree is
			// the destination repo and always carries the full config, with lock
			// targets already repo-root-relative to match the prefixed patch paths.
			repoAnalysis, err := runner.AnalyseRepo(devWorktreePath, &runner.AnalyseRepoOptions{
				Phases: []string{"lock"},

				SourcePaths: scmFiles,
			})
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

		// Upstream paths are relative to the upstream root; prefix them with
		// the directory holding UPSTREAM.ndscm.ts so files land under it. `^`
		// prepends the replacement; `$` is escaped since MoveDiff expands it.
		upstreamPathPrefix := regexp.MustCompile(`^`)
		upstreamPathReplace := strings.ReplaceAll(upstreamDir+"/", "$", "$$")

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

			if modifiers != nil {
				if len(modifiers.includePatterns) > 0 {
					patch.PickDiff(modifiers.includePatterns, modifiers.includePatterns)
				}
				if len(modifiers.excludePatterns) > 0 {
					patch.DropDiff(modifiers.excludePatterns, modifiers.excludePatterns)
				}
				for _, renamePattern := range modifiers.renamePatterns {
					patch.MoveDiff(
						renamePattern.match,
						renamePattern.match,
						renamePattern.replace,
						renamePattern.replace,
					)
				}

				if patch.DiffCount() == 0 {
					seedlog.Warnf("Dropping excluded commit: %v", commitId)
					continue
				}
			}

			// Prefix before dropping locks so the patch paths are
			// repo-root-relative, matching the lock patterns.
			if upstreamDir != "." {
				patch.MoveDiff(upstreamPathPrefix, upstreamPathPrefix, upstreamPathReplace, upstreamPathReplace)
			}

			// Matching both the old and new path drops a diff only when both
			// sides are lock files, sparing renames into/out of a lock path.
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
