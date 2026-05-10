package runner

import (
	"encoding/json"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/configloader"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type RunTask struct {
	// DirPath is the directory to run the command in.
	DirPath string

	// BazelTargets lists bazel targets whose outputs this command needs.
	BazelTargets []string

	// Cmd is the command to run.
	Cmd string
}

// Watcher represents a single target produced by a build rule.
// When a rule produces multiple targets, the first target carries Watch
// and Run; the remaining targets set Together to the first target's name,
// indicating they are produced by the same invocation.
type Watcher struct {
	// Watch lists the source files that matched the rule's watch patterns.
	Watch []string

	// Run lists the commands to execute.
	Run []RunTask

	// Together references the primary target name when this target is a
	// secondary output of the same rule. Watch and Run are empty when
	// Together is set.
	Together string
}

// RepoPhase holds the matched targets for a single build phase within a
// directory (e.g. "build", "test").
type RepoPhase struct {
	// Prerequisites lists the phases that must be up-to-date before this phase can run.
	Prerequisites []string

	// Targets maps target file paths to their build tasks.
	Targets map[string][]Watcher
}

func (rp RepoPhase) String() string {
	b, _ := json.MarshalIndent(rp, "", "  ")
	return string(b)
}

type RepoAnalysis struct {
	Phases map[string]RepoPhase
}

func (ra RepoAnalysis) String() string {
	b, _ := json.MarshalIndent(ra, "", "  ")
	return string(b)
}

func collectRelPaths(dir string, scmFilePaths []string) []string {
	relPaths := []string{}
	for _, fp := range scmFilePaths {
		rel, err := filepath.Rel(dir, fp)
		if err != nil || strings.HasPrefix(rel, "..") {
			continue
		}
		relPaths = append(relPaths, rel)
	}
	sort.Strings(relPaths)
	return relPaths
}

func matchWatch(patterns configloader.StringOrStrings, relPaths []string) ([]string, error) {
	compiled := []*regexp.Regexp{}
	for _, pattern := range patterns {
		re, err := regexp.Compile(pattern)
		if err != nil {
			return nil, seederr.WrapErrorf("invalid regexp: %s", pattern)
		}
		compiled = append(compiled, re)
	}
	matches := []string{}
	for _, relPath := range relPaths {
		for _, re := range compiled {
			if re.MatchString(relPath) {
				matches = append(matches, relPath)
				break
			}
		}
	}
	return matches, nil
}

func prefixPaths(dir string, paths []string) []string {
	result := make([]string, len(paths))
	for i, p := range paths {
		result[i] = filepath.Join(dir, p)
	}
	return result
}

func analyseDir(
	scmDirPath string, dirConfig *configloader.DirConfig, phases []string, scmFilePaths []string,
) (map[string]RepoPhase, error) {
	relPaths := collectRelPaths(scmDirPath, scmFilePaths)
	result := map[string]RepoPhase{}
	for _, phase := range phases {
		prerequisites := []string{}
		// The phases are forced to run sequentially for now to avoid breaking the bazel managed tools for each phase.
		switch phase {
		case "format":
			prerequisites = []string{}
		case "vendor":
			prerequisites = []string{"format"}
		case "bootstrap":
			prerequisites = []string{"format", "vendor"}
		case "tidy":
			prerequisites = []string{"format", "vendor", "bootstrap"}
		case "lock":
			prerequisites = []string{"format", "vendor", "bootstrap", "tidy"}
		case "build":
			prerequisites = []string{"format", "vendor", "bootstrap", "tidy", "lock"}
		case "test":
			prerequisites = []string{"format", "vendor", "bootstrap", "tidy", "lock", "build"}
		}

		rules := map[string]configloader.WatchGenerateRule{}
		switch phase {
		case "format":
			rules = dirConfig.Format
		case "vendor":
			rules = dirConfig.Vendor
		case "bootstrap":
			rules = dirConfig.Bootstrap
		case "tidy":
			rules = dirConfig.Tidy
		case "lock":
			rules = dirConfig.Lock
		case "build":
			rules = dirConfig.Build
		case "test":
			rules = dirConfig.Test
		}

		phaseTargets := map[string][]Watcher{}
		ruleKeys := []string{}
		for ruleKey := range rules {
			ruleKeys = append(ruleKeys, ruleKey)
		}
		sort.Strings(ruleKeys)

		for _, ruleKey := range ruleKeys {
			rule := rules[ruleKey]
			matches, err := matchWatch(rule.Watch, relPaths)
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			if len(matches) == 0 {
				continue
			}
			runTasks := make([]RunTask, len(rule.Run))
			for i, cmd := range rule.Run {
				runTasks[i] = RunTask{
					DirPath:      scmDirPath,
					BazelTargets: []string(rule.NeedBazelBuild),
					Cmd:          cmd,
				}
			}
			if phase == "format" {
				for _, relPath := range matches {
					scmTargetFilePath := filepath.Join(scmDirPath, relPath)
					phaseTargets[scmTargetFilePath] = append(phaseTargets[scmTargetFilePath],
						Watcher{
							Watch: []string{scmTargetFilePath},
							Run:   runTasks,
						})
				}
				continue
			}
			if phase == "test" {
				target := "test_" + ruleKey
				phaseTargets[target] = append(phaseTargets[target], Watcher{
					Watch: prefixPaths(scmDirPath, matches),
					Run:   runTasks,
				})
				continue
			}
			ruleTargets := prefixPaths(scmDirPath, []string(rule.Target))
			if len(ruleTargets) == 0 {
				continue
			}
			phaseTargets[ruleTargets[0]] = append(phaseTargets[ruleTargets[0]], Watcher{
				Watch: append([]string(rule.WatchRepo), prefixPaths(scmDirPath, matches)...),
				Run:   runTasks,
			})
			for _, t := range ruleTargets[1:] {
				phaseTargets[t] = append(phaseTargets[t], Watcher{
					Together: ruleTargets[0],
				})
			}
		}
		if len(phaseTargets) == 0 {
			continue
		}
		result[phase] = RepoPhase{
			Prerequisites: prerequisites,
			Targets:       phaseTargets,
		}
	}
	return result, nil
}

func mergeRepoPhase(basePhase RepoPhase, newPhase RepoPhase) RepoPhase {
	if len(basePhase.Targets) == 0 {
		return newPhase
	}
	prereqSet := map[string]bool{}
	for _, p := range basePhase.Prerequisites {
		prereqSet[p] = true
	}
	for _, p := range newPhase.Prerequisites {
		prereqSet[p] = true
	}
	prerequisites := []string{}
	for p := range prereqSet {
		prerequisites = append(prerequisites, p)
	}
	sort.Strings(prerequisites)

	merged := RepoPhase{
		Prerequisites: prerequisites,
		Targets:       map[string][]Watcher{},
	}
	for k, v := range basePhase.Targets {
		merged.Targets[k] = v
	}
	for k, v := range newPhase.Targets {
		merged.Targets[k] = append(merged.Targets[k], v...)
	}
	return merged
}

func AnalyseRepo(worktree string, phases []string, scmFilePaths []string) (*RepoAnalysis, error) {
	dirConfigs, err := configloader.LoadDirConfigs(worktree, scmFilePaths)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	scmDirPhases := map[string]map[string]RepoPhase{}
	scmDirPaths := []string{}
	for dirConfigPath, dirConfig := range dirConfigs {
		scmDirPath := filepath.Dir(dirConfigPath)
		dirPhases, err := analyseDir(scmDirPath, dirConfig, phases, scmFilePaths)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		scmDirPhases[scmDirPath] = dirPhases
		scmDirPaths = append(scmDirPaths, scmDirPath)
	}

	// sort scmDirPaths with BFS so that parent directories are processed before their children
	sort.Slice(scmDirPaths, func(i, j int) bool {
		di := strings.Count(scmDirPaths[i], string(filepath.Separator))
		dj := strings.Count(scmDirPaths[j], string(filepath.Separator))
		if di != dj {
			return di < dj
		}
		return scmDirPaths[i] < scmDirPaths[j]
	})

	result := &RepoAnalysis{Phases: map[string]RepoPhase{}}
	for _, scmDirPath := range scmDirPaths {
		dirPhases := scmDirPhases[scmDirPath]
		for phase, dirPhase := range dirPhases {
			result.Phases[phase] = mergeRepoPhase(result.Phases[phase], dirPhase)
		}
	}

	return result, nil
}
