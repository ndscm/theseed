package clientcore

import (
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdSubmitOptions struct {
	Force  bool
	Remote string

	Owner string

	MaxCommits int

	FeatureName string
	CutPoint    string
}

// ndSubmitBranchOptions describes how to materialize a submit branch: the suffix naming the
// branch, the start point to root it at, the upstream it tracks, and the commit
// range (applyFrom, applyTo] to replay onto it.
type ndSubmitBranchOptions struct {
	remote string

	owner string

	maxCommits int

	suffix string

	startPoint string
	tracking   string

	applyFrom string
	applyTo   string
}

// ndSubmitBranch materializes the submit branch described by plan: it clears any
// stale worktree/branch, roots the submit branch, replays the planned commit
// range (capped by MaxCommits), signs off, pushes, and tears the worktree down.
func ndSubmitBranch(
	scmProvider scm.Provider, repo *scm.WorkingRepo, options *ndSubmitBranchOptions,
) error {
	// The submit branch is "<owner>/submit/<suffix>", defaulting the owner to the
	// current user when --owner is not given.
	owner := options.owner
	if owner == "" {
		currentUserHandle, err := user.CurrentUserHandle()
		if err != nil {
			return seederr.Wrap(err)
		}
		owner = currentUserHandle
	}
	submitBranch := owner + "/submit/" + options.suffix

	worktreePath := filepath.Join(repo.MonorepoHome, submitBranch)
	_, err := os.Stat(worktreePath)
	if err != nil && !os.IsNotExist(err) {
		return seederr.Wrap(err)
	}
	if err == nil {
		seedlog.Warnf("Worktree %v already exists, removing...", worktreePath)
		err = scmProvider.RemoveWorktree(repo, submitBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	_, err = scmProvider.GetCommitId(submitBranch)
	if err == nil {
		seedlog.Warnf("Branch %v already exists, removing...", submitBranch)
		err = scmProvider.DeleteBranch(repo, submitBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	err = scmProvider.CreateBranch(submitBranch, options.startPoint, options.tracking)
	if err != nil {
		return seederr.Wrap(err)
	}
	worktreePath, err = scmProvider.CreateWorktree(repo, submitBranch)
	if err != nil {
		return seederr.Wrap(err)
	}

	submitPoint := options.applyTo
	if options.maxCommits > 0 {
		commitIds, err := scmProvider.ListCommitIds(options.applyFrom, options.applyTo)
		if err != nil {
			return seederr.Wrap(err)
		}
		if options.maxCommits < len(commitIds) {
			submitPoint = commitIds[options.maxCommits-1]
			seedlog.Warnf("Limiting submit branch to the first %d of %d commits (stopping at %v)",
				options.maxCommits, len(commitIds), submitPoint)
		}
	}

	err = scmProvider.ApplyCommitRange(worktreePath, options.applyFrom, submitPoint)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.SignOff(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.PushBranch(submitBranch, options.remote, submitBranch, true)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.RemoveWorktree(repo, submitBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.DeleteBranch(repo, submitBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// ndSubmitChange submits a change branch selected from a dev worktree. A dev
// branch is a chain of change branches, so it requires a feature name to pick the
// change branch, and optionally cuts a fresh change branch first.
func ndSubmitChange(
	scmProvider scm.Provider, repo *scm.WorkingRepo, devBranch string, options NdSubmitOptions,
) error {
	if options.FeatureName == "" {
		return seederr.WrapErrorf("feature name is required when submitting a dev branch")
	}
	if options.CutPoint != "" {
		err := NdCut(scmProvider, NdCutOptions{
			Force:       options.Force,
			FeatureName: options.FeatureName,
			CutPoint:    options.CutPoint,
		})
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	changeBranch, err := scm.GetChangeBranchName(devBranch, options.FeatureName)
	if err != nil {
		return seederr.Wrap(err)
	}
	baseBranch, err := scm.GetBaseBranchName(devBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	baseTracking, err := scmProvider.GetBranchTracking(baseBranch)
	if err != nil {
		return seederr.WrapErrorf("tracking upstream is missing for %v", baseBranch)
	}
	trackingBranch, err := scmProvider.GetBranchTracking(changeBranch)
	if err != nil {
		return seederr.WrapErrorf("tracking upstream is missing for %v", changeBranch)
	}
	mergeBaseCommitId, err := scmProvider.GetMergeBaseCommitId(baseTracking, changeBranch)
	if err != nil {
		return seederr.WrapErrorf("merge base is missing for %v", changeBranch)
	}
	plan := &ndSubmitBranchOptions{
		remote: options.Remote,

		owner: options.Owner,

		maxCommits: options.MaxCommits,

		suffix: options.FeatureName,

		startPoint: mergeBaseCommitId,
		tracking:   baseTracking,

		applyFrom: trackingBranch,
		applyTo:   changeBranch,
	}
	err = ndSubmitBranch(scmProvider, repo, plan)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// ndSubmitMelt submits a melt worktree. A melt branch is a single linear branch
// (no chain) based directly on the mainline, so it takes no feature name or cut
// point. Its base must already appear on the mainline, and the base..melt range
// must be linear.
func ndSubmitMelt(
	scmProvider scm.Provider, repo *scm.WorkingRepo, meltBranch string, options NdSubmitOptions,
) error {
	if options.FeatureName != "" {
		return seederr.WrapErrorf("feature name must be empty when submitting a melt branch")
	}
	if options.CutPoint != "" {
		return seederr.WrapErrorf("cut point must be empty when submitting a melt branch")
	}
	_, _, upstreamName, err := scm.ParseCanonicalBranch(meltBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	baseBranch, err := scm.GetBaseBranchName(meltBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	baseTracking, err := scmProvider.GetBranchTracking(baseBranch)
	if err != nil {
		return seederr.WrapErrorf("tracking upstream is missing for %v", baseBranch)
	}
	basePoint, err := scmProvider.GetCommitId(baseBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	// The melt base must already be part of the mainline; otherwise the submit
	// branch would carry base commits that never landed on the trunk.
	mergeBase, err := scmProvider.GetMergeBaseCommitId(baseTracking, baseBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	if mergeBase != basePoint {
		return seederr.WrapErrorf(
			"melt base %v (%v) does not appear on %v", baseBranch, basePoint, baseTracking)
	}
	// A melt branch must be a single linear branch; ListCommitIds fails if the
	// base/melt..melt range is not pure (e.g. contains a merge commit).
	_, err = scmProvider.ListCommitIds(baseBranch, meltBranch)
	if err != nil {
		return seederr.WrapErrorf("melt branch %v is not linear: %w", meltBranch, err)
	}
	plan := &ndSubmitBranchOptions{
		remote: options.Remote,

		owner: options.Owner,

		maxCommits: options.MaxCommits,

		suffix: "melt/" + upstreamName,

		startPoint: basePoint,
		tracking:   baseTracking,

		applyFrom: baseBranch,
		applyTo:   meltBranch,
	}
	err = ndSubmitBranch(scmProvider, repo, plan)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func NdSubmit(scmProvider scm.Provider, options NdSubmitOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-submit should not run with --shell-eval")
	}

	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	repo := &scm.WorkingRepo{MonorepoHome: monorepoHome}
	err = scmProvider.QuickVerifyMonorepo(repo)
	if err != nil {
		return seederr.Wrap(err)
	}

	worktreeBranch, err := scmProvider.GetWorktreeBranch("")
	if err != nil {
		return seederr.Wrap(err)
	}

	switch {
	case scm.IsBranchType(worktreeBranch, "dev"):
		err = ndSubmitChange(scmProvider, repo, worktreeBranch, options)
	case scm.IsBranchType(worktreeBranch, "melt"):
		err = ndSubmitMelt(scmProvider, repo, worktreeBranch, options)
	default:
		return seederr.WrapErrorf("current worktree %v is neither a dev nor a melt branch", worktreeBranch)
	}
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
