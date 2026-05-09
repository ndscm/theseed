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

func formatFile(
	worktreeAbsPath string, scmTargetPath string, watchers []Watcher, bazelGround *BazelGround,
) error {
	targetAbsPath := filepath.Join(worktreeAbsPath, scmTargetPath)
	escapedTarget := escapeFilePathForBash(targetAbsPath)
	for _, watcher := range watchers {
		for _, runTask := range watcher.Run {
			dirAbsPath := filepath.Join(worktreeAbsPath, runTask.DirPath)
			bashCmd := runTask.Cmd
			bashCmd = strings.ReplaceAll(bashCmd, "{{TARGET}}", escapedTarget)
			bashCmd, err := bazelGround.fulfill(bashCmd, runTask.BazelTargets)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = seedshell.ImpureOptionsRun(
				[]seedshell.RunOption{func(cmd *exec.Cmd) {
					cmd.Dir = dirAbsPath

					// Comfort the aspect js_binary rule
					cmd.Env = append(cmd.Env, "BAZEL_BINDIR=.")
				}},
				"bash", "-c", bashCmd)
			if err != nil {
				return seederr.Wrap(err)
			}
		}
	}
	return nil
}

func formatFiles(worktree string, formatPhase RepoPhase, filePaths []string) error {
	seedlog.Debugf("Formatting files:\n%v", strings.Join(filePaths, "\n"))
	worktreeAbsPath, err := filepath.Abs(worktree)
	if err != nil {
		return seederr.Wrap(err)
	}

	bazelGround := NewBazelGround()
	err = bazelGround.collect(formatPhase)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = bazelGround.build(worktree)
	if err != nil {
		return seederr.Wrap(err)
	}

	fileSet := map[string]bool{}
	for _, p := range filePaths {
		fileSet[p] = true
	}

	scmTargetPaths := []string{}
	for k := range formatPhase.Targets {
		scmTargetPaths = append(scmTargetPaths, k)
	}
	sort.Strings(scmTargetPaths)

	errGroup := errgroup.Group{}
	errGroup.SetLimit(100)
	errsMutex := sync.Mutex{}
	errs := []error{}

	for _, scmTargetPath := range scmTargetPaths {
		watchers := formatPhase.Targets[scmTargetPath]
		hasDirty := false
		for _, watcher := range watchers {
			for _, watching := range watcher.Watch {
				if fileSet[watching] {
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
		errGroup.Go(func() error {
			err := formatFile(worktreeAbsPath, scmTargetPath, watchers, bazelGround)
			if err != nil {
				errsMutex.Lock()
				defer errsMutex.Unlock()
				errs = append(errs, err)
			}
			return nil
		})
	}
	errGroup.Wait()

	seedlog.Infof("Finished formatting files: count=%d", len(scmTargetPaths))
	return errors.Join(errs...)
}

func FormatDirtyFiles(worktree string, scmFilePaths []string, scmDirtyPaths []string) error {
	repoAnalysis, err := AnalyseRepo(worktree, []string{"format"}, scmFilePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Debugf("Analysis:\n%v", repoAnalysis)

	formatPhase, ok := repoAnalysis.Phases["format"]
	if !ok {
		return nil
	}
	err = formatFiles(worktree, formatPhase, scmDirtyPaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func FormatAllFiles(worktree string, scmFilePaths []string) error {
	repoAnalysis, err := AnalyseRepo(worktree, []string{"format"}, scmFilePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Debugf("Analysis:\n%v", repoAnalysis)

	formatPhase, ok := repoAnalysis.Phases["format"]
	if !ok {
		return nil
	}
	err = formatFiles(worktree, formatPhase, scmFilePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
