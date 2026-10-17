package runner

import (
	"crypto/sha256"
	"encoding/json/jsontext"
	"encoding/json/v2"
	"path/filepath"
	"regexp"
	"sort"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/configloader"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagUseBazelGround = seedflag.DefineString("use_bazel_ground", "", "Use bazel ground to optimize bazel command execution.")

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

	// Passive controls what triggers the watcher to run. Normally a watcher runs
	// when its stamp is missing or when its watch list is updated. When Passive is
	// set, a missing stamp alone does not trigger a run; only an updated watch list
	// does. This is typically useful for formatter watchers.
	Passive bool `json:"passive"`

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

	UseBazelGround bool

	// Watchers lists the matched build rules for this phase.
	Watchers []Watcher
}

func (rp RepoPhase) String() string {
	b, _ := json.Marshal(rp, jsontext.WithIndent("  "))
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
	b, _ := json.Marshal(ra, jsontext.WithIndent("  "))
	return string(b)
}

func collectRelPaths(dir string, sourcePaths []string) []string {
	relPaths := []string{}
	for _, sourcePath := range sourcePaths {
		relPath, err := filepath.Rel(dir, sourcePath)
		if err != nil || strings.HasPrefix(relPath, "..") {
			continue
		}
		relPaths = append(relPaths, relPath)
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

type analyseDirOptions struct {
	Phases []string

	SourcePaths []string
}

func analyseDir(
	scmDirPath string, dirConfig *configloader.DirConfig, opts *analyseDirOptions,
) (map[string]RepoPhase, error) {
	if opts == nil {
		opts = &analyseDirOptions{}
	}
	relPaths := collectRelPaths(scmDirPath, opts.SourcePaths)
	result := map[string]RepoPhase{}
	for _, phase := range opts.Phases {
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

			runTasks := []RunTask{}
			for _, candidateBuildSystem := range BuildSystems() {
				switch candidateBuildSystem {
				case "bazel":
					if rule.Bazel != nil && len(rule.Bazel.Run) > 0 {
						for _, cmd := range rule.Bazel.Run {
							runTasks = append(runTasks, RunTask{
								DirPath:      scmDirPath,
								BazelTargets: []string(rule.Bazel.Build),
								Cmd:          cmd,
							})
						}
					}
				case "buck":
					if rule.Buck != nil && len(rule.Buck.Run) > 0 {
						for _, cmd := range rule.Buck.Run {
							runTasks = append(runTasks, RunTask{
								DirPath: scmDirPath,
								Cmd:     cmd,
							})
						}
					}
				}
				if len(runTasks) > 0 {
					break
				}
			}
			if len(runTasks) == 0 && len(rule.Run) > 0 {
				for _, cmd := range rule.Run {
					runTasks = append(runTasks, RunTask{
						DirPath: scmDirPath,
						Cmd:     cmd,
					})
				}
			}
			if len(runTasks) == 0 {
				continue
			}

			if phase == "format" {
				for _, relPath := range matches {
					scmTargetFilePath := filepath.Join(scmDirPath, relPath)
					watchers = append(watchers, Watcher{
						Phase:   phase,
						Passive: true,
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

type AnalyseRepoOptions struct {
	Phases []string

	DirConfigPaths []string

	SourcePaths []string
}

func AnalyseRepo(worktreePath string, opts *AnalyseRepoOptions) (*RepoAnalysis, error) {
	if opts == nil {
		opts = &AnalyseRepoOptions{}
	}

	dirConfigPaths := opts.DirConfigPaths
	if len(dirConfigPaths) == 0 {
		dirConfigPaths = opts.SourcePaths
	}
	dirConfigs, err := configloader.LoadDirConfigs(worktreePath, dirConfigPaths)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	scmDirPhases := map[string]map[string]RepoPhase{}
	scmDirPaths := []string{}
	for dirConfigPath, dirConfig := range dirConfigs {
		scmDirPath := filepath.Dir(dirConfigPath)
		dirPhases, err := analyseDir(scmDirPath, dirConfig, &analyseDirOptions{
			Phases: opts.Phases,

			SourcePaths: opts.SourcePaths,
		})
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

	bazelGroundPhases := strings.Split(flagUseBazelGround.Get(), ",")
	for _, phase := range opts.Phases {
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

		useBazelGround := false
		for _, p := range bazelGroundPhases {
			if p == phase {
				useBazelGround = true
				break
			}
		}

		result.Phases[phase] = RepoPhase{
			Prerequisites:  prerequisites,
			UseBazelGround: useBazelGround,
			Watchers:       result.Phases[phase].Watchers,
		}
	}

	return result, nil
}

func CareAnalysis(repoAnalysis *RepoAnalysis, careRepoPaths []string) (*RepoAnalysis, error) {
	dependentWatchers := map[string][]*Watcher{}
	for _, repoPhase := range repoAnalysis.Phases {
		for i := range repoPhase.Watchers {
			watcher := &repoPhase.Watchers[i]
			for _, watchRepoPath := range watcher.Watch {
				dependentWatchers[watchRepoPath] = append(dependentWatchers[watchRepoPath], watcher)
			}
		}
	}

	fifo := []string{}
	for _, p := range careRepoPaths {
		fifo = append(fifo, p)
	}
	careSet := map[string]bool{}
	for len(fifo) > 0 {
		repoPath := fifo[0]
		fifo = fifo[1:]
		if !careSet[repoPath] {
			careSet[repoPath] = true
			for _, watcher := range dependentWatchers[repoPath] {
				for _, targetRepoPath := range watcher.Targets {
					fifo = append(fifo, targetRepoPath)
				}
			}
		}
	}

	careAnalysis := &RepoAnalysis{
		Phases: map[string]RepoPhase{},
	}
	for phase, repoPhase := range repoAnalysis.Phases {
		phaseWatchers := []Watcher{}
		for _, watcher := range repoPhase.Watchers {
			keep := false
			for _, watchingRepoPath := range watcher.Watch {
				if careSet[watchingRepoPath] {
					keep = true
					break
				}
			}
			if keep {
				phaseWatchers = append(phaseWatchers, watcher)
			}
		}
		careAnalysis.Phases[phase] = RepoPhase{
			Prerequisites:  repoPhase.Prerequisites,
			UseBazelGround: repoPhase.UseBazelGround,
			Watchers:       phaseWatchers,
		}
	}
	return careAnalysis, nil
}
