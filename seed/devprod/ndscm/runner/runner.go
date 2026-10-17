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

	all bool

	repoStamper *RepoStamper

	bazelGround *BazelGround

	watchingMap map[string][]*Watcher

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

func (r *Runner) runWatchingMap() (map[string]bool, error) {
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
		for _, watchers := range r.watchingMap {
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
					stamped, changed, err := r.repoStamper.CheckStamp(watcher, stampGroupWatch, watching)
					if err != nil {
						return nil, seederr.Wrap(err)
					}
					if !changed {
						continue
					}
					// A passive watcher runs only when its watch list is
					// updated, not when its stamp is missing. A missing stamp
					// is exactly the changed-but-not-stamped case, so skip it.
					// With all set, stamps are reset, so run regardless.
					if watcher.Passive && !stamped && !r.all {
						continue
					}
					hasDirty = true
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

	seedlog.Infof("Triggered watchers: count=%d", finalCount)
	return phaseDirtySet, nil
}

func (r *Runner) collectPhase(repoPhase *RepoPhase) error {
	// Watchers from every phase are accumulated into a single shared bazel
	// ground and can run intermixed, so "use bazel ground" is a run-wide
	// property rather than a per-phase one: it is only safe when every
	// participating phase opts in. As soon as one phase opts out we disable the
	// bazel ground for the rest of the run, and we never re-enable it. The
	// disable is intentionally one-way (there is no setUse(true) path) — a later
	// opt-in phase cannot undo an earlier phase's incompatibility.
	if !repoPhase.UseBazelGround {
		r.bazelGround.setUse(false)
	}
	err := r.bazelGround.collect(*repoPhase)
	if err != nil {
		return seederr.Wrap(err)
	}

	if r.watchingMap == nil {
		r.watchingMap = map[string][]*Watcher{}
	}
	phaseWatchingMap := generateWatchingMap(repoPhase)
	for target, watchers := range phaseWatchingMap {
		r.watchingMap[target] = append(r.watchingMap[target], watchers...)
	}
	return nil
}

type RunOptions struct {
	Phases []string

	All bool

	DirConfigPaths []string

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
	r.all = opts.All

	repoAnalysis, err := AnalyseRepo(r.worktreePath, &AnalyseRepoOptions{
		Phases: opts.Phases,

		DirConfigPaths: opts.DirConfigPaths,

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
		repoAnalysis = careAnalysis
		r.all = true
	}

	if r.all {
		err := r.repoStamper.RemoveStamps(repoAnalysis)
		if err != nil {
			return seederr.Wrap(err)
		}
	}

	r.bazelGround = NewBazelGround(r.worktreePath, true)
	r.targetLocks = &sync.Map{}
	for _, phase := range r.phases {
		repoPhase, ok := repoAnalysis.Phases[phase]
		if !ok {
			continue
		}
		err := r.collectPhase(&repoPhase)
		if err != nil {
			return seederr.Wrap(err)
		}

		skipPhase := true
		for _, watcher := range repoPhase.Watchers {
			for _, watching := range watcher.Watch {
				stamped, changed, err := r.repoStamper.CheckStamp(&watcher, stampGroupWatch, watching)
				if err != nil {
					return seederr.Wrap(err)
				}
				if !changed {
					continue
				}
				// A passive watcher is not triggered by a missing stamp alone,
				// so it must not keep the phase from being skipped. With all
				// set, stamps are reset, so the phase must still run.
				if watcher.Passive && !stamped && !r.all {
					continue
				}
				skipPhase = false
				break
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
		err = r.bazelGround.build(r.worktreePath)
		if err != nil {
			return seederr.Wrap(err)
		}
		_, err = r.runWatchingMap()
		if err != nil {
			return seederr.Wrap(err)
		}
		seedlog.Infof("Finished phase: phase=%q", phase)
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
