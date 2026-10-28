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

func CheckTestable(scmProvider scm.Provider, commit string, belong *string) (bool, error) {
	belongTo := ""
	if belong != nil {
		belongTo = *belong
	} else {
		belongTo = flagBelong.Get()
	}
	if belongTo == "" {
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

	if belongTo == "auto" {
		monorepoHome, err := scm.MonorepoHome()
		if err != nil {
			return false, seederr.Wrap(err)
		}
		repo := &scm.WorkingRepo{MonorepoHome: monorepoHome}
		worktreeName, _, err := scmProvider.GetCurrentWorktree(repo)
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
			// currentBranch is the trunk, a branch without a tracking branch that
			// may itself be a remote branch such as origin/main. The delta walk
			// above never inspects its own [root, currentBranch] range, so check
			// whether the target is reachable from it before concluding the commit
			// has no belong branch (e.g. a detached HEAD during a rebase).
			mergeBaseCommitId, err := scmProvider.GetMergeBaseCommitId(targetCommitId, currentBranch)
			if err == nil && mergeBaseCommitId == targetCommitId {
				found = true
			}
		}
		if found {
			belongTo = currentBranch
		} else {
			seedlog.Warnf("Target is not found on worktree branch. target=%v, commit=%v, worktree=%v", targetCommit, targetCommitId, worktreeName)
			belongTo = ""
		}
	}
	seedlog.Debugf("Checking testable. target=%v belong=%v", targetCommitId, belongTo)

	testable := true
	if belongTo != "" {
		belongTrackingBranch, err := scmProvider.GetBranchTracking(belongTo)
		if err != nil {
			belongTrackingBranch = ""
		}
		belongBranchCommitIds, err := scmProvider.ListCommitIds(belongTrackingBranch, belongTo)
		if err != nil {
			return false, seederr.Wrap(err)
		}

		// A commit is clean when it carries no side-effect and is not itself a break
		// (i.e. no later commit points a side-effect back at it). Clean commits are a
		// consistent, testable checkpoint regardless of any open range above them.
		targetHasSideEffect := false
		for _, extended := range targetCommitMetadata.Extended {
			if extended.Key == "side-effect-of-change-uuid" {
				targetHasSideEffect = true
				break
			}
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
			hasSideEffect := false
			for _, extended := range commitMetadata.Extended {
				if extended.Key == "side-effect-of-change-uuid" {
					hasSideEffect = true
					untestableRange[extended.Value] = commitMetadata.ChangeUuid
				}
			}
			// A commit is a break when a later side-effect already pointed back at it,
			// recorded as its change uuid sitting in untestableRange. A commit that is
			// neither a side-effect nor a break is clean and resets the range so the
			// untestable range stays continuous: a side-effect pointing back to an
			// ancient break must not drag the clean commits in between into the range.
			_, isBreak := untestableRange[commitMetadata.ChangeUuid]
			if !hasSideEffect && !isBreak {
				untestableRange = map[string]string{}
				continue
			}
			delete(untestableRange, commitMetadata.ChangeUuid)
		}
		if !found {
			return false, seederr.WrapErrorf("target is not found on belong branch. target=%v uid=%v belong=%v", targetCommit, targetCommitId, belongTo)
		}
		_, targetIsBreak := untestableRange[targetCommitMetadata.ChangeUuid]
		testable = (!targetHasSideEffect && !targetIsBreak) || len(untestableRange) == 0
		if !testable {
			seedlog.Infof("Commit is not testable. commit=%v, untestable=%v", targetCommitId, untestableRange)
		}
	}
	return testable, nil
}
