package runner

import (
	"errors"
	"os/exec"
	"path/filepath"
	"sort"
	"strings"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
	"golang.org/x/sync/errgroup"
)

var bashEscapeReplacer = strings.NewReplacer(
	`\`, `\\`,
	`"`, `\"`,
	`$`, `\$`,
	"`", "\\`",
)

func escapeFilePathForBash(filePath string) string {
	return bashEscapeReplacer.Replace(filePath)
}

// formatFile runs all watcher commands against a single target file.
func formatFile(
	worktreePath string, targetRepoPath string, watchers []Watcher, bazelGround *BazelGround,
) error {
	targetAbsPath := filepath.Join(worktreePath, targetRepoPath)
	escapedTarget := escapeFilePathForBash(targetAbsPath)
	for _, watcher := range watchers {
		for _, runTask := range watcher.Run {
			dirAbsPath := filepath.Join(worktreePath, runTask.DirPath)
			bashCmd := runTask.Cmd
			bashCmd = strings.ReplaceAll(bashCmd, "{{TARGET}}", escapedTarget)
			bashCmd, err := bazelGround.fulfill(bashCmd, runTask.BazelTargets)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = seedshell.ImpureOptionsRun(
				[]seedshell.RunOption{func(cmd *exec.Cmd) {
					cmd.Dir = dirAbsPath
				}},
				"bash", "-c", bashCmd)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
	}
	return nil
}

// formatFiles runs formatters on all targets that have at least one dirty watched file.
func formatFiles(worktreePath string, formatPhase RepoPhase, dirtyRepoPaths []string) error {
	seedlog.Debugf("Formatting files:\n%v", strings.Join(dirtyRepoPaths, "\n"))

	bazelGround := NewBazelGround()
	err := bazelGround.collect(formatPhase)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = bazelGround.build(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}

	dirtySet := map[string]bool{}
	for _, p := range dirtyRepoPaths {
		dirtySet[p] = true
	}

	targetRepoPaths := []string{}
	for k := range formatPhase.Targets {
		targetRepoPaths = append(targetRepoPaths, k)
	}
	sort.Strings(targetRepoPaths)

	errGroup := errgroup.Group{}
	errGroup.SetLimit(100)
	errsMutex := sync.Mutex{}
	errs := []error{}

	finalCount := 0
	for _, targetRepoPath := range targetRepoPaths {
		watchers := formatPhase.Targets[targetRepoPath]
		hasDirty := false
		for _, watcher := range watchers {
			for _, watching := range watcher.Watch {
				if dirtySet[watching] {
					hasDirty = true
					break
				}
			}
			if hasDirty {
				break
			}
		}
		if !hasDirty {
			continue
		}
		finalCount++
		errGroup.Go(func() error {
			err := formatFile(worktreePath, targetRepoPath, watchers, bazelGround)
			if err != nil {
				errsMutex.Lock()
				defer errsMutex.Unlock()
				errs = append(errs, err)
			}
			return nil
		})
	}
	errGroup.Wait()

	seedlog.Infof("Finished formatting files: count=%d", finalCount)
	return errors.Join(errs...)
}

// FormatDirtyFiles formats only the targets whose watched files appear in
// scmDirtyPaths.
func FormatDirtyFiles(worktreePath string, scmFilePaths []string, scmDirtyPaths []string) error {
	repoAnalysis, err := AnalyseRepo(worktreePath, []string{"format"}, scmFilePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Debugf("Analysis:\n%v", repoAnalysis)

	formatPhase, ok := repoAnalysis.Phases["format"]
	if !ok {
		return nil
	}
	err = formatFiles(worktreePath, formatPhase, scmDirtyPaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// FormatAllFiles formats all targets in the repository, treating every file as dirty.
func FormatAllFiles(worktreePath string, scmFilePaths []string) error {
	repoAnalysis, err := AnalyseRepo(worktreePath, []string{"format"}, scmFilePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Debugf("Analysis:\n%v", repoAnalysis)

	formatPhase, ok := repoAnalysis.Phases["format"]
	if !ok {
		return nil
	}
	err = formatFiles(worktreePath, formatPhase, scmFilePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
