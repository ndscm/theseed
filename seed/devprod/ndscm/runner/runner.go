package runner

import (
	"encoding/json/jsontext"
	"encoding/json/v2"
	"errors"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
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
	workerCount int

	worktreePath string

	phases []string

	repoStamper *RepoStamper

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
		bashCmd, err := r.bazelGround.fulfill(bashCmd, runTask.DirPath, runTask.BazelTargets)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		runEnv := map[string]string{}
		runEnv["BUILD_WORKSPACE_DIRECTORY"] = r.worktreePath
		runEnv["BUILD_WORKING_DIRECTORY"] = filepath.Join(r.worktreePath, runTask.DirPath)
		err = seedshell.ImpureOptionsRun(
			[]seedshell.RunOption{func(cmd *exec.Cmd) {
				cmd.Dir = dirAbsPath
				cmd.Env = os.Environ()
				for k, v := range runEnv {
					cmd.Env = append(cmd.Env, k+"="+v)
				}
			}},
			"bash", "-c", bashCmd)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}

	for _, watchRepoPath := range watcher.Watch {
		_, err := r.repoStamper.Stamp(watcher, stampGroupWatch, watchRepoPath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}

	watcherDirtySet := map[string]bool{}
	for _, targetRepoPath := range watcher.Targets {
		changed, err := r.repoStamper.Stamp(watcher, stampGroupTarget, targetRepoPath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if changed {
			watcherDirtySet[targetRepoPath] = true
		}
	}
	return watcherDirtySet, nil
}

func (r *Runner) runPhase(repoPhase *RepoPhase) (map[string]bool, error) {
	r.bazelGround = NewBazelGround(r.worktreePath, repoPhase.UseBazelGround)
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
	errGroup.SetLimit(r.workerCount)
	errsMutex := sync.Mutex{}
	errs := []error{}

	phaseDirtySet := map[string]bool{}
	finalCount := 0
	round := 0
	for {
		round++
		seen := map[*Watcher]bool{}
		roundDirty := sync.Map{}
		for _, watchers := range watching {
			for _, watcher := range watchers {
				if seen[watcher] {
					continue
				}
				ready := true
				for _, watching := range watcher.Watch {
					_, err := os.Stat(filepath.Join(r.worktreePath, watching))
					if err != nil {
						if !os.IsNotExist(err) {
							return nil, seederr.Wrap(err)
						}
						ready = false
						break
					}
				}
				if !ready {
					continue
				}

				hasDirty := false
				for _, watching := range watcher.Watch {
					changed, err := r.repoStamper.CheckChanged(watcher, stampGroupWatch, watching)
					if err != nil {
						return nil, seederr.Wrap(err)
					}
					if changed {
						hasDirty = true
					}
				}
				if !hasDirty {
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
						roundDirty.Store(changedRepoPath, true)
					}
					return nil
				})
			}
		}
		errGroup.Wait()
		if len(errs) > 0 {
			return nil, errors.Join(errs...)
		}
		roundDirtySet := map[string]bool{}
		roundDirty.Range(func(key any, value any) bool {
			p := key.(string)
			phaseDirtySet[p] = true
			roundDirtySet[p] = true
			return true
		})
		if len(roundDirtySet) == 0 {
			break
		}
		seedlog.Infof("Finished round %d: dirty=%v", round, roundDirtySet)
	}

	seedlog.Infof("Finished phase: count=%d", finalCount)
	return phaseDirtySet, nil
}

type RunOptions struct {
	Phases []string

	SourcePaths []string

	CarePaths []string
}

func (r *Runner) Run(opts *RunOptions) error {
	if opts == nil {
		opts = &RunOptions{}
	}
	if len(r.phases) > 0 {
		return seederr.WrapErrorf("phases have already been set, cannot run again")
	}
	if len(opts.Phases) == 0 {
		return nil
	}

	repoAnalysis, err := AnalyseRepo(r.worktreePath, &AnalyseRepoOptions{
		Phases: opts.Phases,

		SourcePaths: opts.SourcePaths,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	analysisJson, err := json.Marshal(repoAnalysis, jsontext.WithIndent("  "))
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
	r.repoStamper = NewRepoStamper(r.worktreePath)

	if len(opts.CarePaths) > 0 {
		careAnalysis, err := CareAnalysis(repoAnalysis, opts.CarePaths)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = r.repoStamper.RemoveStamps(careAnalysis)
		if err != nil {
			return seederr.Wrap(err)
		}
		repoAnalysis = careAnalysis
	}

	for _, phase := range r.phases {
		repoPhase, ok := repoAnalysis.Phases[phase]
		if !ok {
			continue
		}

		skipPhase := true
		for _, watcher := range repoPhase.Watchers {
			for _, watching := range watcher.Watch {
				changed, err := r.repoStamper.CheckChanged(&watcher, stampGroupWatch, watching)
				if err != nil {
					return seederr.Wrap(err)
				}
				if changed {
					skipPhase = false
					break
				}
			}
			if !skipPhase {
				break
			}
		}
		if skipPhase {
			seedlog.Infof("Skipping phase: phase=%q", phase)
			continue
		}

		seedlog.Infof("Started phase: phase=%q", phase)
		_, err := r.runPhase(&repoPhase)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

type RunnerOptions struct {
	WorkerCount int
}

func CreateRunner(
	worktreePath string, opts *RunnerOptions,
) (*Runner, error) {
	if opts == nil {
		opts = &RunnerOptions{}
	}
	if opts.WorkerCount == 0 {
		opts.WorkerCount = runtime.NumCPU()
	}

	r := &Runner{
		workerCount:  opts.WorkerCount,
		worktreePath: worktreePath,
	}
	return r, nil
}
