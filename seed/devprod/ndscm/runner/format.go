package runner

import (
	"os/exec"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
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

func formatFiles(worktree string, formatPhase RepoPhase, filePaths []string) error {
	seedlog.Debugf("Formatting files:\n%v", strings.Join(filePaths, "\n"))
	worktreeAbsPath, err := filepath.Abs(worktree)
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
		if hasDirty {
			targetAbsPath := filepath.Join(worktreeAbsPath, scmTargetPath)
			escapedTarget := escapeFilePathForBash(targetAbsPath)
			for _, watcher := range watchers {
				for _, runTask := range watcher.Run {
					dirAbsPath := filepath.Join(worktreeAbsPath, runTask.DirPath)
					bashCmd := runTask.Cmd
					bashCmd = strings.ReplaceAll(bashCmd, "{{TARGET}}", escapedTarget)
					err := seedshell.ImpureOptionsRun(
						[]seedshell.RunOption{func(cmd *exec.Cmd) {
							cmd.Dir = dirAbsPath
						}},
						"bash", "-c", bashCmd)
					if err != nil {
						return seederr.Wrap(err)
					}
				}
			}
		}
	}
	return nil
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
