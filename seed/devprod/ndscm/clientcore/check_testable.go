package clientcore

import (
	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagBelong = seedflag.DefineString("belong", "auto",
	"Belong branch for testable check. If set to auto, it will automatically find the belong branch on dev branch. If set to empty, the check will be disabled.",
)

func CheckTestable(scmProvider scm.Provider, commit string) (bool, error) {
	belong := flagBelong.Get()
	if belong == "" {
		seedlog.Warnf("Testable check is disabled as belong is set to empty.")
		return true, nil
	}

	targetCommit := commit
	if targetCommit == "" {
		targetCommit = "HEAD"
	}
	targetCommitId, err := scmProvider.GetCommitId(targetCommit)
	if err != nil {
		return false, seederr.Wrap(err)
	}
	targetCommitMetadata, err := scmProvider.GetCommitMetadata(targetCommitId)
	if err != nil {
		return false, seederr.Wrap(err)
	}

	if belong == "auto" {
		monorepoHome, err := scm.MonorepoHome()
		if err != nil {
			return false, seederr.Wrap(err)
		}
		worktreeName, _, err := scmProvider.GetCurrentWorktree(monorepoHome)
		if err != nil {
			return false, seederr.Wrap(err)
		}

		currentBranch := worktreeName
		found := false
		for {
			currentTrackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
			if err != nil {
				break
			}
			branchCommitIds, err := scmProvider.ListCommitIds(currentTrackingBranch, currentBranch)
			if err != nil {
				return false, seederr.Wrap(err)
			}
			for _, commitId := range branchCommitIds {
				commitMetadata, err := scmProvider.GetCommitMetadata(commitId)
				if err != nil {
					return false, seederr.Wrap(err)
				}
				if targetCommitId == commitId ||
					(targetCommitMetadata.ChangeUuid != "" &&
						targetCommitMetadata.ChangeUuid == commitMetadata.ChangeUuid) {
					found = true
					break
				}
			}
			if found {
				break
			}
			currentBranch = currentTrackingBranch
		}
		if !found {
			seedlog.Warnf("Target is not found on worktree branch, guess on the trunk. target=%v, commit=%v, worktree=%v", targetCommit, targetCommitId, worktreeName)
		}
		belong = currentBranch
	}
	seedlog.Debugf("Checking testable. target=%v belong=%v", targetCommitId, belong)

	belongTrackingBranch, err := scmProvider.GetBranchTracking(belong)
	if err != nil {
		belongTrackingBranch = ""
	}
	belongBranchCommitIds, err := scmProvider.ListCommitIds(belongTrackingBranch, belong)
	if err != nil {
		return false, seederr.Wrap(err)
	}

	found := false
	untestableRange := map[string]string{}
	for i := len(belongBranchCommitIds) - 1; i >= 0; i-- {
		commitId := belongBranchCommitIds[i]
		commitMetadata, err := scmProvider.GetCommitMetadata(commitId)
		if err != nil {
			return false, seederr.Wrap(err)
		}
		seedlog.Debugf("Checked untestable status. commit=%v change=%v untestable=%v",
			commitId, commitMetadata.ChangeUuid, untestableRange,
		)
		if targetCommitId == commitId ||
			(targetCommitMetadata.ChangeUuid != "" &&
				targetCommitMetadata.ChangeUuid == commitMetadata.ChangeUuid) {
			found = true
			break
		}
		for _, extended := range commitMetadata.Extended {
			if extended.Key == "side-effect-of-change-uuid" {
				untestableRange[extended.Value] = commitMetadata.ChangeUuid
			}
		}
		delete(untestableRange, commitMetadata.ChangeUuid)
	}
	if !found {
		return false, seederr.WrapErrorf("target is not found on belong branch. target=%v uid=%v belong=%v", targetCommit, targetCommitId, belong)
	}
	testable := len(untestableRange) == 0
	if !testable {
		seedlog.Infof("Commit is not testable. commit=%v, untestable=%v", targetCommitId, untestableRange)
	}
	return testable, nil
}
