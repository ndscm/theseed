package runner

import (
	"errors"
	"os/exec"
	"path/filepath"
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

// formatFile runs a watcher's commands
func formatFile(
	worktreePath string, watcher Watcher, bazelGround *BazelGround, targetLocks *sync.Map,
) error {
	if len(watcher.Targets) == 0 {
		return nil
	}

	for _, targetRepoPath := range watcher.Targets {
		// Defensive retrive all locks before length check.
		targetLock, _ := targetLocks.LoadOrStore(targetRepoPath, &sync.Mutex{})
		mu := targetLock.(*sync.Mutex)
		mu.Lock()
		defer mu.Unlock()
	}

	if len(watcher.Targets) > 1 {
		return seederr.WrapErrorf("format task should not contain multiple targets")
	}

	targetAbsPath := filepath.Join(worktreePath, watcher.Targets[0])
	escapedTarget := escapeFilePathForBash(targetAbsPath)
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

	targetLocks := sync.Map{}
	errGroup := errgroup.Group{}
	errGroup.SetLimit(100)
	errsMutex := sync.Mutex{}
	errs := []error{}

	finalCount := 0
	for _, watcher := range formatPhase.Watchers {
		hasDirty := false
		for _, watching := range watcher.Watch {
			if dirtySet[watching] {
				hasDirty = true
				break
			}
		}
		if !hasDirty {
			continue
		}
		finalCount++
		errGroup.Go(func() error {
			err := formatFile(worktreePath, watcher, bazelGround, &targetLocks)
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
