package runner

import (
	"encoding/json"
	"os"
	"path/filepath"
	"sort"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func escapeMakeTarget(s string) (string, error) {
	if strings.ContainsAny(s, ": \t\n") {
		return "", seederr.WrapErrorf("invalid make target: %q", s)
	}
	result := strings.ReplaceAll(s, "$", "$$")
	return result, nil
}

func generatePhaseBlock(repoAnalysis *RepoAnalysis, phase string) (string, error) {
	if phase == "format" {
		return "format:\n\tndscm format\n\n", nil
	}
	dirPhase, ok := repoAnalysis.Phases[phase]
	if !ok {
		return "", nil
	}
	if len(dirPhase.Targets) == 0 {
		return "", nil
	}

	targetKeys := make([]string, 0, len(dirPhase.Targets))
	for k := range dirPhase.Targets {
		targetKeys = append(targetKeys, k)
	}
	sort.Strings(targetKeys)

	allTargets := []string{}
	rules := strings.Builder{}
	for _, target := range targetKeys {
		tasks := dirPhase.Targets[target]
		if len(tasks) == 0 {
			continue
		}
		if tasks[0].Together != "" {
			continue
		}
		ruleTargets := []string{target}
		for _, other := range targetKeys {
			otherTasks := dirPhase.Targets[other]
			if len(otherTasks) > 0 && otherTasks[0].Together == target {
				ruleTargets = append(ruleTargets, other)
			}
		}
		allTargets = append(allTargets, ruleTargets...)

		escapedTargets := []string{}
		for _, t := range ruleTargets {
			escapedTarget, err := escapeMakeTarget(t)
			if err != nil {
				return "", seederr.Wrap(err)
			}
			escapedTargets = append(escapedTargets, escapedTarget)
		}
		rules.WriteString(strings.Join(escapedTargets, " "))
		rules.WriteString(":")
		for _, task := range tasks {
			for _, w := range task.Watch {
				rules.WriteString(" ")
				escapedWatch, err := escapeMakeTarget(w)
				if err != nil {
					return "", seederr.Wrap(err)
				}
				rules.WriteString(escapedWatch)
			}
		}
		rules.WriteString("\n")
		for _, task := range tasks {
			for _, runTask := range task.Run {
				bashCmd := runTask.Cmd
				bashCmd = strings.ReplaceAll(bashCmd, "$", "$$")
				bashCmd = strings.ReplaceAll(bashCmd, "'", `'\''`)
				rules.WriteString("\tcd ")
				escapedDirPath, err := escapeMakeTarget(runTask.DirPath)
				if err != nil {
					return "", seederr.Wrap(err)
				}
				rules.WriteString(escapedDirPath)
				rules.WriteString(" && bash -c '")
				rules.WriteString(bashCmd)
				rules.WriteString("'\n")
			}
		}
		rules.WriteString("\n")
	}

	result := strings.Builder{}
	result.WriteString(phase)
	result.WriteString(":")
	for _, p := range dirPhase.Prerequisites {
		if _, ok := repoAnalysis.Phases[p]; !ok {
			continue
		}
		result.WriteString(" ")
		result.WriteString(p)
	}
	for _, t := range allTargets {
		result.WriteString(" ")
		escapedTarget, err := escapeMakeTarget(t)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		result.WriteString(escapedTarget)
	}
	result.WriteString("\n\n")
	result.WriteString(rules.String())

	return result.String(), nil
}

func generateMakefile(repoAnalysis *RepoAnalysis, phases []string) (string, error) {
	phaseBlocks := map[string]string{}
	for _, phase := range phases {
		phaseBlock, err := generatePhaseBlock(repoAnalysis, phase)
		if err != nil {
			return "", seederr.Wrap(err)
		}
		if phaseBlock != "" {
			phaseBlocks[phase] = phaseBlock
		}
	}

	final := strings.Builder{}
	final.WriteString("SHELL := /usr/bin/env bash\n\n")
	final.WriteString(".PHONY: all")
	for _, phase := range phases {
		_, ok := phaseBlocks[phase]
		if ok {
			final.WriteString(" ")
			final.WriteString(phase)
		}
	}
	final.WriteString("\n\nall:")
	for _, phase := range phases {
		_, ok := phaseBlocks[phase]
		if ok {
			final.WriteString(" ")
			final.WriteString(phase)
		}
	}
	final.WriteString("\n\n")

	for _, phase := range phases {
		block, ok := phaseBlocks[phase]
		if ok {
			final.WriteString(block)
		}
	}

	return final.String(), nil
}

func GenerateMakefile(worktreePath string, scmFilePaths []string) (string, error) {
	phases := []string{
		"format",
		"vendor",
		"bootstrap",
		"tidy",
		"lock",
		"build",
		"test",
	}
	repoAnalysis, err := AnalyseRepo(worktreePath, phases, scmFilePaths)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	analysisJson, err := json.MarshalIndent(repoAnalysis, "", "  ")
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = os.MkdirAll(filepath.Join(worktreePath, ".cache/ndscm"), 0755)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	err = os.WriteFile(filepath.Join(worktreePath, ".cache/ndscm/analysis.json"), analysisJson, 0644)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	content, err := generateMakefile(repoAnalysis, phases)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return content, nil
}
