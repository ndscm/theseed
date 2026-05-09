package runner

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type BazelTargetInfo struct {
	ArtifactPaths   []string
	ExecutablePaths []string
}

func (info BazelTargetInfo) artifacts() (string, error) {
	if len(info.ArtifactPaths) == 0 {
		return "", seederr.WrapErrorf("no artifacts")
	}
	return strings.Join(info.ArtifactPaths, " "), nil
}

func (info BazelTargetInfo) executable() (string, error) {
	if len(info.ExecutablePaths) == 0 {
		return "", seederr.WrapErrorf("no executables")
	}
	if len(info.ExecutablePaths) > 1 {
		return "", seederr.WrapErrorf("multiple executables found: %v", info.ExecutablePaths)
	}
	return info.ExecutablePaths[0], nil
}

type BazelGround struct {
	targetMapMutex sync.RWMutex
	targetMap      map[string]BazelTargetInfo

	builtMutex sync.Mutex
	built      bool
}

func (gnd *BazelGround) collect(phase RepoPhase) error {
	gnd.builtMutex.Lock()
	defer gnd.builtMutex.Unlock()

	if gnd.targetMap == nil {
		return seederr.WrapErrorf("bazel ground not initialized")
	}
	if gnd.built {
		return seederr.WrapErrorf("collect after build is not allowed")
	}

	gnd.targetMapMutex.Lock()
	defer gnd.targetMapMutex.Unlock()

	for _, watchers := range phase.Targets {
		for _, watcher := range watchers {
			for _, runTask := range watcher.Run {
				for _, bazelTarget := range runTask.BazelTargets {
					if _, ok := gnd.targetMap[bazelTarget]; !ok {
						gnd.targetMap[bazelTarget] = BazelTargetInfo{}
					}
				}
			}
		}
	}
	return nil
}

var bazelCqueryStarlarkExpr = strings.Join([]string{
	`"ARTIFACTS"`,
	`+ "\n"`,
	`+ "\n".join([f.path for f in target.files.to_list()])`,
	`+ "\n"`,
	`+ "EXECUTABLES"`,
	`+ "\n"`,
	`+ (target.files_to_run.executable.path if target.files_to_run.executable else "")`,
}, " ")

func parseBazelCqueryOutput(worktreeAbsPath string, output string) BazelTargetInfo {
	info := BazelTargetInfo{}
	sections := strings.SplitN(output, "EXECUTABLES\n", 2)

	artifactSection := ""
	executableSection := ""
	if len(sections) == 2 {
		artifactSection = sections[0]
		executableSection = sections[1]
	}

	artifactSection = strings.TrimPrefix(artifactSection, "ARTIFACTS\n")
	for _, line := range strings.Split(strings.TrimSpace(artifactSection), "\n") {
		if line != "" {
			info.ArtifactPaths = append(info.ArtifactPaths, filepath.Join(worktreeAbsPath, line))
		}
	}
	for _, line := range strings.Split(strings.TrimSpace(executableSection), "\n") {
		if line != "" {
			info.ExecutablePaths = append(info.ExecutablePaths, filepath.Join(worktreeAbsPath, line))
		}
	}
	return info
}

func (gnd *BazelGround) build(worktree string) error {
	gnd.builtMutex.Lock()
	defer gnd.builtMutex.Unlock()

	if gnd.built {
		return seederr.WrapErrorf("build bazel tools twice is not allowed")
	}

	worktreeAbsPath, err := filepath.Abs(worktree)
	if err != nil {
		return seederr.Wrap(err)
	}

	gnd.targetMapMutex.Lock()
	defer gnd.targetMapMutex.Unlock()

	bazelTargets := []string{}
	for bazelTarget := range gnd.targetMap {
		bazelTargets = append(bazelTargets, bazelTarget)
	}
	sort.Strings(bazelTargets)

	if len(bazelTargets) == 0 {
		gnd.built = true
		return nil
	}

	err = seedshell.ImpureOptionsRun(
		[]seedshell.RunOption{func(cmd *exec.Cmd) {
			cmd.Dir = worktreeAbsPath
		}},
		"bazel", append([]string{"build"}, bazelTargets...)...)
	if err != nil {
		return seederr.Wrap(err)
	}

	// TODO(nagi): Parse BEP output to get artifact and executable paths.

	for _, bazelTarget := range bazelTargets {
		output, err := seedshell.PureOptionsOutput(
			[]seedshell.RunOption{func(cmd *exec.Cmd) {
				cmd.Dir = worktreeAbsPath
			}},
			"bazel", "cquery", "--output=starlark",
			"--starlark:expr="+bazelCqueryStarlarkExpr,
			bazelTarget)
		if err != nil {
			return seederr.Wrap(err)
		}
		info := parseBazelCqueryOutput(worktreeAbsPath, string(output))
		seedlog.Debugf("Built bazel target: target=%s artifacts=%v executables=%v",
			bazelTarget, info.ArtifactPaths, info.ExecutablePaths)
		gnd.targetMap[bazelTarget] = info
	}

	gnd.built = true
	return nil
}

func (gnd *BazelGround) get(bazelTarget string) (BazelTargetInfo, error) {
	gnd.targetMapMutex.RLock()
	defer gnd.targetMapMutex.RUnlock()

	info, ok := gnd.targetMap[bazelTarget]
	if !ok {
		return BazelTargetInfo{}, seederr.WrapErrorf("bazel target not collected: %s", bazelTarget)
	}
	return info, nil
}

func (gnd *BazelGround) fulfill(bashCmd string, bazelTargets []string) (string, error) {
	for i, bazelTarget := range bazelTargets {
		targetInfo, err := gnd.get(bazelTarget)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		if i == 0 {
			if strings.Contains(bashCmd, "{{BAZEL_RUN}}") {
				executablePath, err := targetInfo.executable()
				if err != nil {
					return "", seederr.Wrap(err)
				}
				bashCmd = strings.ReplaceAll(bashCmd, "{{BAZEL_RUN}}", executablePath)
			}
			if strings.Contains(bashCmd, "{{BAZEL_BUILD}}") {
				artifacts, err := targetInfo.artifacts()
				if err != nil {
					return "", seederr.Wrap(err)
				}
				bashCmd = strings.ReplaceAll(bashCmd, "{{BAZEL_BUILD}}", artifacts)
			}
		}
		if strings.Contains(bashCmd, "{{BAZEL_RUN["+strconv.Itoa(i)+"]}}") {
			executablePath, err := targetInfo.executable()
			if err != nil {
				return "", seederr.Wrap(err)
			}
			bashCmd = strings.ReplaceAll(bashCmd, "{{BAZEL_RUN["+strconv.Itoa(i)+"]}}", executablePath)
		}
		if strings.Contains(bashCmd, "{{BAZEL_BUILD["+strconv.Itoa(i)+"]}}") {
			artifacts, err := targetInfo.artifacts()
			if err != nil {
				return "", seederr.Wrap(err)
			}
			bashCmd = strings.ReplaceAll(bashCmd, "{{BAZEL_BUILD["+strconv.Itoa(i)+"]}}", artifacts)
		}
	}
	return bashCmd, nil
}

func NewBazelGround() *BazelGround {
	return &BazelGround{
		targetMap: map[string]BazelTargetInfo{},
	}
}
