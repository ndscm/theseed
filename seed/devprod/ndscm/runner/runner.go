package runner

import (
	"encoding/json"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"slices"
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

func generateWatchingMap(repoPhase *RepoPhase) map[string][]*Watcher {
	watching := map[string][]*Watcher{}
	for i := range repoPhase.Watchers {
		watcher := &repoPhase.Watchers[i]
		for _, target := range watcher.Targets {
			watching[target] = append(watching[target], watcher)
		}
	}
	return watching
}

type Runner struct {
	worktreePath string

	scmFilePaths []string

	phases []string

	bazelGround *BazelGround

	targetLocks *sync.Map
}

func (r *Runner) runWatcher(watcher *Watcher) (map[string]bool, error) {
	if len(watcher.Targets) == 0 {
		return nil, nil
	}

	sortedTargets := slices.Sorted(slices.Values(watcher.Targets))
	for _, targetRepoPath := range sortedTargets {
		targetLock, _ := r.targetLocks.LoadOrStore(targetRepoPath, &sync.Mutex{})
		mu := targetLock.(*sync.Mutex)
		mu.Lock()
		defer mu.Unlock()
	}

	for _, runTask := range watcher.Run {
		dirAbsPath := filepath.Join(r.worktreePath, runTask.DirPath)
		bashCmd := runTask.Cmd
		if watcher.Phase == "format" {
			if len(watcher.Targets) > 1 {
				return nil, seederr.WrapErrorf("format task should not contain multiple targets")
			}
			targetAbsPath := filepath.Join(r.worktreePath, watcher.Targets[0])
			escapedTarget := escapeFilePathForBash(targetAbsPath)
			bashCmd = strings.ReplaceAll(bashCmd, "{{TARGET}}", escapedTarget)
		}
		bashCmd, err := r.bazelGround.fulfill(bashCmd, runTask.BazelTargets)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		err = seedshell.ImpureOptionsRun(
			[]seedshell.RunOption{func(cmd *exec.Cmd) {
				cmd.Dir = dirAbsPath
			}},
			"bash", "-c", bashCmd)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}

	watcherDirtySet := map[string]bool{}
	for _, targetRepoPath := range watcher.Targets {
		changed, err := Stamp(r.worktreePath, watcher, targetRepoPath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if changed {
			watcherDirtySet[targetRepoPath] = true
		}
	}
	return watcherDirtySet, nil
}

func (r *Runner) runPhase(repoPhase *RepoPhase, dirtySet map[string]bool) (map[string]bool, error) {
	r.bazelGround = NewBazelGround()
	err := r.bazelGround.collect(*repoPhase)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = r.bazelGround.build(r.worktreePath)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	r.targetLocks = &sync.Map{}
	watching := generateWatchingMap(repoPhase)

	errGroup := errgroup.Group{}
	errGroup.SetLimit(100)
	errsMutex := sync.Mutex{}
	errs := []error{}

	phaseDirtySet := map[string]bool{}
	finalCount := 0
	round := 0
	for {
		round++
		seen := map[*Watcher]bool{}
		newDirtySet := sync.Map{}
		for _, watchers := range watching {
			for _, watcher := range watchers {
				if seen[watcher] {
					continue
				}
				ready := true
				hasDirty := false
				for _, watching := range watcher.Watch {
					_, err := os.Stat(filepath.Join(r.worktreePath, watching))
					if err != nil {
						if !os.IsNotExist(err) {
							return nil, seederr.Wrap(err)
						}
						ready = false
						break
					}
					if dirtySet[watching] {
						hasDirty = true
					}
				}
				if !ready || !hasDirty {
					continue
				}

				seen[watcher] = true
				finalCount++

				errGroup.Go(func() error {
					watcherDirtySet, err := r.runWatcher(watcher)
					if err != nil {
						// Log the error immediately for better visibility.
						seedlog.Errorf("Run watcher failed: watcher=%v err=%v", watcher, err)
						errsMutex.Lock()
						defer errsMutex.Unlock()
						errs = append(errs, err)
						return nil
					}
					for changedRepoPath := range watcherDirtySet {
						newDirtySet.Store(changedRepoPath, true)
					}
					return nil
				})
			}
		}
		errGroup.Wait()
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
		dirtySet = map[string]bool{}
		newDirtySet.Range(func(key any, value any) bool {
			p := key.(string)
			dirtySet[p] = true
			phaseDirtySet[p] = true
			return true
		})
		if len(dirtySet) == 0 {
			break
		}
		seedlog.Infof("Finished round %d: dirty=%v", round, dirtySet)
	}

	seedlog.Infof("Finished phase: count=%d", finalCount)
	return phaseDirtySet, nil
}

func (r *Runner) Run(phases []string, dirtyRepoPaths []string) error {
	if len(r.phases) > 0 {
		return seederr.WrapErrorf("phases have already been set, cannot run again")
	}
	if len(phases) == 0 {
		return nil
	}

	repoAnalysis, err := AnalyseRepo(r.worktreePath, phases, r.scmFilePaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	analysisJson, err := json.MarshalIndent(repoAnalysis, "", "  ")
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.MkdirAll(filepath.Join(r.worktreePath, ".cache/ndscm"), 0755)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(filepath.Join(r.worktreePath, ".cache/ndscm/analysis.json"), analysisJson, 0644)
	if err != nil {
		return seederr.Wrap(err)
	}
	sortedPhases, err := repoAnalysis.TopologicalSort()
	if err != nil {
		return seederr.Wrap(err)
	}
	r.phases = sortedPhases

	dirtySet := map[string]bool{}
	if len(dirtyRepoPaths) > 0 {
		for _, p := range dirtyRepoPaths {
			dirtySet[p] = true
		}
		_, err := StampAll(r.worktreePath, repoAnalysis)
		if err != nil {
			return seederr.Wrap(err)
		}
	} else {
		stampDirtySet, err := StampAll(r.worktreePath, repoAnalysis)
		if err != nil {
			return seederr.Wrap(err)
		}
		dirtySet = stampDirtySet
	}

	for _, phase := range r.phases {
		repoPhase, ok := repoAnalysis.Phases[phase]
		if !ok {
			continue
		}
		seedlog.Infof("Started phase: phase=%q dirtyCount=%d", phase, len(dirtySet))
		seedlog.Debugf("Started phase: dirty=%v", dirtySet)
		phaseDirtySet, err := r.runPhase(&repoPhase, dirtySet)
		if err != nil {
			return seederr.Wrap(err)
		}
		for repoPath := range phaseDirtySet {
			dirtySet[repoPath] = true
		}
	}
	return nil
}

func (r *Runner) Format(dirtyRepoPaths []string) error {
	err := r.Run([]string{"format"}, dirtyRepoPaths)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func CreateRunner(
	worktreePath string, scmFilePaths []string,
) (*Runner, error) {
	r := &Runner{
		worktreePath: worktreePath,
		scmFilePaths: scmFilePaths,
	}
	return r, nil
}
