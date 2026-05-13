package runner

import (
	_ "embed"
	"os"
	"os/exec"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"sync"

	"github.com/ndscm/theseed/seed/devprod/rbe/bes/go/bep"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

//go:embed default_executable_aspect.bzl
var defaultExecutableAspect string

func prepareDefaultExecutableAspect(worktreePath string) error {
	err := os.MkdirAll(filepath.Join(worktreePath, ".cache/ndscm"), 0755)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(filepath.Join(worktreePath, ".cache/ndscm/BUILD.bazel"), []byte{}, 0644)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = os.WriteFile(
		filepath.Join(worktreePath, ".cache/ndscm/default_executable_aspect.bzl"),
		[]byte(defaultExecutableAspect), 0644)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// See: https://github.com/bazelbuild/bazel/blob/8.6.0/src/main/java/com/google/devtools/build/lib/cmdline/TargetPattern.java#L911
func absolutizeBazelTarget(dirRepoPath string, target string) string {
	if strings.HasPrefix(target, "//") {
		return target
	}
	if strings.HasPrefix(target, "@") {
		return target
	}
	if strings.HasPrefix(target, ":") {
		return "//" + dirRepoPath + target
	}
	if dirRepoPath == "" {
		return "//" + target
	}
	return "//" + dirRepoPath + "/" + target
}

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
	worktreePath string

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

	for _, watcher := range phase.Watchers {
		for _, runTask := range watcher.Run {
			for _, bazelTarget := range runTask.BazelTargets {
				absoluteBazelTarget := absolutizeBazelTarget(runTask.DirPath, bazelTarget)
				_, ok := gnd.targetMap[absoluteBazelTarget]
				if !ok {
					gnd.targetMap[absoluteBazelTarget] = BazelTargetInfo{}
				}
			}
		}
	}
	return nil
}

func (gnd *BazelGround) build(worktreePath string) error {
	gnd.builtMutex.Lock()
	defer gnd.builtMutex.Unlock()

	if gnd.built {
		return seederr.WrapErrorf("build bazel tools twice is not allowed")
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

	err := prepareDefaultExecutableAspect(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	bazelArgs := []string{
		"build",
		"--aspects=.cache/ndscm/default_executable_aspect.bzl%default_executable_aspect",
		"--output_groups=+default_executable",
	}
	err = seedshell.ImpureOptionsRun(
		[]seedshell.RunOption{func(cmd *exec.Cmd) {
			cmd.Dir = worktreePath
		}},
		"bazel", append(bazelArgs, bazelTargets...)...)
	if err != nil {
		return seederr.Wrap(err)
	}

	bepPath := filepath.Join(worktreePath, ".bep")
	data, err := os.ReadFile(bepPath)
	if err != nil {
		return seederr.Wrap(err)
	}
	bepEvents, err := bep.ParseBuildEventProtos(data)
	if err != nil {
		return seederr.Wrap(err)
	}

	for _, bazelTarget := range bazelTargets {
		info := BazelTargetInfo{}
		artifacts, err := bep.QueryOutput(bepEvents, bazelTarget, "default")
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, f := range artifacts {
			pathParts := append(f.GetPathPrefix(), f.GetName())
			info.ArtifactPaths = append(info.ArtifactPaths, filepath.Join(worktreePath, filepath.Join(pathParts...)))
		}
		executables, err := bep.QueryOutput(bepEvents, bazelTarget, "default_executable")
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, f := range executables {
			pathParts := append(f.GetPathPrefix(), f.GetName())
			info.ExecutablePaths = append(info.ExecutablePaths, filepath.Join(worktreePath, filepath.Join(pathParts...)))
		}
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

func (gnd *BazelGround) fulfill(bashCmd string, dirRepoPath string, bazelTargets []string) (string, error) {
	for i, bazelTarget := range bazelTargets {
		absoluteBazelTarget := absolutizeBazelTarget(dirRepoPath, bazelTarget)
		targetInfo, err := gnd.get(absoluteBazelTarget)
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

func NewBazelGround(worktreePath string) *BazelGround {
	return &BazelGround{
		worktreePath: worktreePath,
		targetMap:    map[string]BazelTargetInfo{},
	}
}
