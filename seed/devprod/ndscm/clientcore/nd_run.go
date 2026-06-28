package clientcore

import (
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/runner"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func scanDirConfigs(worktreePath string, careRepoPaths []string) []string {
	seen := map[string]bool{}
	dirConfigRepoPaths := []string{}
	for _, careRepoPath := range careRepoPaths {
		dir := filepath.Dir(careRepoPath)
		for {
			dirConfigRepoPath := filepath.Join(dir, "DIR.ndscm.ts")
			if !seen[dirConfigRepoPath] {
				seen[dirConfigRepoPath] = true
				_, err := os.Stat(filepath.Join(worktreePath, dirConfigRepoPath))
				if err == nil {
					dirConfigRepoPaths = append(dirConfigRepoPaths, dirConfigRepoPath)
				}
			}
			if dir == "." || dir == string(filepath.Separator) {
				break
			}
			dir = filepath.Dir(dir)
		}
	}
	return dirConfigRepoPaths
}

func mergeRepoPaths(slices ...[]string) []string {
	seen := map[string]bool{}
	merged := []string{}
	for _, slice := range slices {
		for _, item := range slice {
			if seen[item] {
				continue
			}
			seen[item] = true
			merged = append(merged, item)
		}
	}
	return merged
}

type NdRunOptions struct {
	Workers int
	All     bool
	Changed bool

	CarePaths   []string
	SkipScmScan bool

	Phases []string
}

func NdRun(scmProvider scm.Provider, options NdRunOptions) error {
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		seedlog.Warnf("Current worktree is not connected by ndscm")
	}
	_, worktreePath, err := scmProvider.GetCurrentWorktree(monorepoHome)
	if err != nil {
		return seederr.Wrap(err)
	}

	careRepoPaths := []string{}
	for _, carePath := range options.CarePaths {
		// Resolve relative care paths against the process CWD so that an
		// interactive run from a subdirectory (e.g. `nd format foo.go`) refers
		// to the file next to the user rather than worktree-root/foo.go.
		absPath, err := filepath.Abs(carePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		careRepoPath, err := filepath.Rel(worktreePath, absPath)
		if err != nil {
			return seederr.Wrap(err)
		}
		if careRepoPath == ".." || strings.HasPrefix(careRepoPath, ".."+string(filepath.Separator)) {
			seedlog.Warnf("Care path %q is outside the worktree %q; skipping", carePath, worktreePath)
			continue
		}
		careRepoPaths = append(careRepoPaths, careRepoPath)
	}

	scmFilePaths := []string{}
	if !options.SkipScmScan {
		scmFiles, err := scmProvider.ListFiles(worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		scmFilePaths = scmFiles
	}

	if len(careRepoPaths) == 0 {
		if options.All {
			careRepoPaths = scmFilePaths
		} else if options.Changed {
			dirtySet := map[string]bool{}
			headChanges, err := scmProvider.ListCommitFiles("HEAD")
			if err != nil {
				return seederr.Wrap(err)
			}
			for _, headFileStatus := range headChanges {
				dirtySet[headFileStatus.To] = true
			}
			dirtyFiles, err := scmProvider.ListDirtyFiles(worktreePath)
			if err != nil {
				return seederr.Wrap(err)
			}
			for _, dirtyFileStatus := range dirtyFiles {
				dirtySet[dirtyFileStatus.From] = false
				dirtySet[dirtyFileStatus.To] = true
			}
			for repoPath, dirty := range dirtySet {
				if dirty {
					careRepoPaths = append(careRepoPaths, repoPath)
				}
			}
		}
	}

	dirConfigRepoPaths := scanDirConfigs(worktreePath, careRepoPaths)
	allRepoPaths := mergeRepoPaths(scmFilePaths, careRepoPaths, dirConfigRepoPaths)
	r, err := runner.CreateRunner(options.Workers, worktreePath, allRepoPaths)
	if err != nil {
		return seederr.Wrap(err)
	}

	err = r.Run(options.Phases, careRepoPaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
