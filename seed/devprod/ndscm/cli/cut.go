package main

import (
	"log"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm/git"
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
	dirtyFiles, err := git.GetStatus("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(dirtyFiles) > 0 {
		return seederr.WrapErrorf("workspace is dirty:\n%v", dirtyFiles)
	}
	devBranch, err := git.GetCurrentBranch("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if !strings.HasPrefix(devBranch, "dev") {
		return seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranch)
	}
	targetCommitHash, err := git.GetCommitHash("", target)
	if err != nil {
		return seederr.Wrap(err)
	}
	currentBranch := devBranch
	for !strings.HasPrefix(currentBranch, "base/") {
		trackingBranch, err := git.GetBranchTracking("", currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if !strings.HasPrefix(trackingBranch, "change/") && !strings.HasPrefix(trackingBranch, "base/") {
			return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", currentBranch, trackingBranch)
		}
		mergeCommits, err := git.ListMergeCommitHash("", trackingBranch, currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if len(mergeCommits) > 0 {
			return seederr.WrapErrorf("current dev branch (%v) is not a pure branch and contains merge commit:\n%v", currentBranch, mergeCommits)
		}
		branchCommits, err := git.ListCommitHash("", trackingBranch, currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		found := false
		for _, commitHash := range branchCommits {
			if targetCommitHash == commitHash {
				found = true
				break
			}
		}
		if found {
			err := git.CreateBranch("", "change/"+featureName, targetCommitHash, trackingBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = git.SetBranchTracking("", currentBranch, "change/"+featureName)
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
