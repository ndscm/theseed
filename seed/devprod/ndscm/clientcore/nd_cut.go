package clientcore

import (
	"log"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

func NdCut(scmProvider scm.Provider, args []string) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-cut should not run with --shell-eval")
	}
	if len(args) != 2 {
		return seederr.WrapErrorf("nd-cut usage: nd cut <feature-name> <ref>|<hash>")
	}
	featureName := strings.TrimSpace(args[0])
	target := strings.TrimSpace(args[1])
	dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(dirtyFiles) > 0 {
		return seederr.WrapErrorf("workspace is dirty:\n%v", dirtyFiles)
	}
	devBranch, err := scmProvider.GetWorktreeBranch("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if !strings.HasPrefix(devBranch, "dev") {
		return seederr.WrapErrorf("workspace branch is not a dev branch: %v", devBranch)
	}
	targetCommitId, err := scmProvider.GetCommitId(target)
	if err != nil {
		return seederr.Wrap(err)
	}
	currentBranch := devBranch
	for !strings.HasPrefix(currentBranch, "base/") {
		trackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if !strings.HasPrefix(trackingBranch, "change/") && !strings.HasPrefix(trackingBranch, "base/") {
			return seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", currentBranch, trackingBranch)
		}
		branchCommits, err := scmProvider.ListCommitIds(trackingBranch, currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		found := false
		for _, commitId := range branchCommits {
			if targetCommitId == commitId {
				found = true
				break
			}
		}
		if found {
			err := scmProvider.CreateBranch("change/"+featureName, targetCommitId, trackingBranch)
			if err != nil {
				return seederr.Wrap(err)
			}
			err = scmProvider.SetBranchTracking(currentBranch, "change/"+featureName)
			if err != nil {
				return seederr.Wrap(err)
			}
			log.Printf("Created change request as %v", "change/"+featureName)
			return nil
		}
		currentBranch = trackingBranch
	}
	return seederr.WrapErrorf("target %v (%v) does not exist on %v branch", target, targetCommitId, devBranch)
}
