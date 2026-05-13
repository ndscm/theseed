package runner

import (
	"crypto/sha256"
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
	DirPath string `json:"dirPath"`

	// BazelTargets lists bazel targets whose outputs this command needs.
	BazelTargets []string `json:"bazelTargets"`

	// Cmd is the command to run.
	Cmd string `json:"cmd"`
}

// Watcher represents a single build rule's matched result.
// Targets lists all output paths produced by the rule invocation.
type Watcher struct {
	Phase string `json:"phase"`

	// Targets lists the output file paths produced by this rule.
	Targets []string `json:"targets"`

	// Watch lists the source files that matched the rule's watch patterns.
	Watch []string `json:"watch"`

	// Run lists the commands to execute.
	Run []RunTask `json:"run"`
}

func (w Watcher) Sha256() ([32]byte, []byte, error) {
	data, err := json.Marshal(w)
	if err != nil {
		return [32]byte{}, nil, seederr.Wrap(err)
	}
	digest := sha256.Sum256(data)
	return digest, data, nil
}

// RepoPhase holds the matched targets for a single build phase within a
// directory (e.g. "build", "test").
type RepoPhase struct {
	// Prerequisites lists the phases that must be up-to-date before this phase can run.
	Prerequisites []string

	// Watchers lists the matched build rules for this phase.
	Watchers []Watcher
}

func (rp RepoPhase) String() string {
	b, _ := json.MarshalIndent(rp, "", "  ")
	return string(b)
}

type RepoAnalysis struct {
	Phases map[string]RepoPhase
}

func (ra RepoAnalysis) TopologicalSort() ([]string, error) {
	inDegreeMap := map[string]int{}
	dependentsMap := map[string][]string{}
	for phase := range ra.Phases {
		if _, ok := inDegreeMap[phase]; !ok {
			inDegreeMap[phase] = 0
		}
		for _, prerequisite := range ra.Phases[phase].Prerequisites {
			dependentsMap[prerequisite] = append(dependentsMap[prerequisite], phase)
			inDegreeMap[phase]++
		}
	}
	fifo := []string{}
	for phase, inDegree := range inDegreeMap {
		if inDegree == 0 {
			fifo = append(fifo, phase)
		}
	}
	result := []string{}
	for len(fifo) > 0 {
		phase := fifo[0]
		fifo = fifo[1:]
		result = append(result, phase)
		for _, dependent := range dependentsMap[phase] {
			inDegreeMap[dependent]--
			if inDegreeMap[dependent] == 0 {
				fifo = append(fifo, dependent)
			}
		}
	}
	if len(result) != len(ra.Phases) {
		return nil, seederr.WrapErrorf("circular dependency detected in phases")
	}
	return result, nil
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

		watchers := []Watcher{}
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
					watchers = append(watchers, Watcher{
						Phase:   phase,
						Targets: []string{scmTargetFilePath},
						Watch:   []string{scmTargetFilePath},
						Run:     runTasks,
					})
				}
				continue
			}
			if phase == "test" {
				target := "test_" + ruleKey
				watchers = append(watchers, Watcher{
					Phase:   phase,
					Targets: []string{target},
					Watch:   prefixPaths(scmDirPath, matches),
					Run:     runTasks,
				})
				continue
			}
			ruleTargets := prefixPaths(scmDirPath, []string(rule.Target))
			if len(ruleTargets) == 0 {
				continue
			}
			watchers = append(watchers, Watcher{
				Phase:   phase,
				Targets: ruleTargets,
				Watch:   append([]string(rule.WatchRepo), prefixPaths(scmDirPath, matches)...),
				Run:     runTasks,
			})
		}
		if len(watchers) == 0 {
			continue
		}
		result[phase] = RepoPhase{
			Watchers: watchers,
		}
	}
	return result, nil
}

func mergeRepoPhase(basePhase RepoPhase, newPhase RepoPhase) RepoPhase {
	if len(basePhase.Watchers) == 0 {
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

	return RepoPhase{
		Prerequisites: prerequisites,
		Watchers:      append(basePhase.Watchers, newPhase.Watchers...),
	}
}

func AnalyseRepo(worktreePath string, phases []string, scmFilePaths []string) (*RepoAnalysis, error) {
	dirConfigs, err := configloader.LoadDirConfigs(worktreePath, scmFilePaths)
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

	for _, phase := range phases {
		// The phases are forced to run sequentially for now to avoid breaking the bazel managed tools for each phase.
		candidates := []string{}
		switch phase {
		case "format":
			candidates = []string{}
		case "vendor":
			candidates = []string{"format"}
		case "bootstrap":
			candidates = []string{"format", "vendor"}
		case "tidy":
			candidates = []string{"format", "vendor", "bootstrap"}
		case "lock":
			candidates = []string{"format", "vendor", "bootstrap", "tidy"}
		case "build":
			candidates = []string{"format", "vendor", "bootstrap", "tidy", "lock"}
		case "test":
			candidates = []string{"format", "vendor", "bootstrap", "tidy", "lock", "build"}
		}
		prerequisites := []string{}
		for _, c := range candidates {
			if len(result.Phases[c].Watchers) > 0 {
				prerequisites = append(prerequisites, c)
			}
		}
		result.Phases[phase] = RepoPhase{
			Prerequisites: prerequisites,
			Watchers:      result.Phases[phase].Watchers,
		}
	}

	return result, nil
}
