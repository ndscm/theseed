package main

import (
	"log"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdCut(args []string, ndConfig *common.NdConfig) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-cut should not run with --shell-eval")
	}
	if len(args) != 3 {
		return seederr.WrapErrorf("nd-cut usage: nd cut <feature-name> <ref>|<hash>")
	}
	featureName := strings.TrimSpace(args[1])
	target := strings.TrimSpace(args[2])
	checkDirtyOutput, err := seedshell.PureOutput("git", "status", "--porcelain")
	if err != nil {
		return seederr.Wrap(err)
	}
	if strings.TrimSpace(string(checkDirtyOutput)) != "" {
		return seederr.WrapErrorf("workspace is dirty:\n%v", string(checkDirtyOutput))
	}
	devBranchOutput, err := seedshell.PureOutput("git", "branch", "--show-current")
	if err != nil {
		return seederr.Wrap(err)
	}
	devBranch := strings.TrimSpace(string(devBranchOutput))
	if !strings.HasPrefix(devBranch, "dev") {
		return seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranch)
	}
	targetCommitHashOutput, err := seedshell.PureOutput("git", "rev-parse", target)
	if err != nil {
		return seederr.Wrap(err)
	}
	targetCommitHash := strings.TrimSpace(string(targetCommitHashOutput))
	currentBranch := devBranch
	for !strings.HasPrefix(currentBranch, "base/") {
		trackingBranchOutput, err := seedshell.PureOutput("git", "rev-parse", "--abbrev-ref", "--symbolic-full-name", currentBranch+"@{upstream}")
		if err != nil {
			return seederr.WrapErrorf("tracking upstream is missing for %v", currentBranch)
		}
		trackingBranch := strings.TrimSpace(string(trackingBranchOutput))
		if !strings.HasPrefix(trackingBranch, "change/") && !strings.HasPrefix(trackingBranch, "base/") {
			return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", currentBranch, trackingBranch)
		}
		mergeCommitsOutput, err := seedshell.PureOutput("git", "rev-list", "--ancestry-path", "--merges", trackingBranch+".."+currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if strings.TrimSpace(string(mergeCommitsOutput)) != "" {
			return seederr.WrapErrorf("current dev branch (%v) is not a pure branch and contains merge commit:\n%v", currentBranch, string(mergeCommitsOutput))
		}
		branchCommitsOutput, err := seedshell.PureOutput("git", "rev-list", "--ancestry-path", trackingBranch+".."+currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		branchCommits := strings.Split(strings.TrimSpace(string(branchCommitsOutput)), "\n")
		for i, commitHash := range branchCommits {
			branchCommits[i] = strings.TrimSpace(commitHash)
		}
		found := false
		for _, commitHash := range branchCommits {
			if targetCommitHash == commitHash {
				found = true
				break
			}
		}
		if found {
			err := seedshell.ImpureRun("git", "branch", "change/"+featureName, targetCommitHash)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = seedshell.ImpureRun("git", "branch", "--set-upstream-to="+trackingBranch, "change/"+featureName)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = seedshell.ImpureRun("git", "branch", "--set-upstream-to=change/"+featureName, currentBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			log.Printf("Created change request as %v", "change/"+featureName)
			return nil
		}
		currentBranch = trackingBranch
	}
	return seederr.WrapErrorf("target %v (%v) does not exist on %v branch", target, targetCommitHash, devBranch)
}
