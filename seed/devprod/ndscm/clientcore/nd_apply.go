package clientcore

import (
	"errors"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdApplyOptions struct {
	Remote string
	Owner  string

	FeatureName string
}

type RemoteChangeBranch struct {
	Remote string

	Owner string

	FeatureName string
}

func (r RemoteChangeBranch) String() string {
	return r.Remote + "/" + r.Owner + "/" + r.FeatureName
}

func listRemoteChangeBranch(allRemoteBranches []string) ([]RemoteChangeBranch, error) {
	remoteChangeBranches := []RemoteChangeBranch{}
	for _, branch := range allRemoteBranches {
		parts := strings.Split(branch, "/")
		if len(parts) != 3 {
			continue
		}
		remoteChangeBranches = append(remoteChangeBranches, RemoteChangeBranch{
			Remote:      parts[0],
			Owner:       parts[1],
			FeatureName: parts[2],
		})
	}
	return remoteChangeBranches, nil
}

func NdApply(scmProvider scm.Provider, options NdApplyOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-apply should not run with --shell-eval")
	}
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.QuickVerifyMonorepo()
	if err != nil {
		return seederr.Wrap(err)
	}
	worktreePath, err := scmProvider.GetCurrentWorktree()
	if err != nil {
		return seederr.Wrap(err)
	}
	devBranch, err := scmProvider.GetBranchWorktreeBranch(monorepoHome, worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	if !scmProvider.IsDevBranch(devBranch) {
		return seederr.WrapErrorf("workspace is not a dev worktree: %v", devBranch)
	}
	dirtyFiles, err := scmProvider.GetWorktreeDirtyFiles("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(dirtyFiles) > 0 {
		return seederr.WrapErrorf("workspace is dirty:\n%v", dirtyFiles)
	}
	operation, err := scmProvider.GetWorktreeOperation("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if operation != "" {
		return seederr.WrapErrorf("an ongoing %v operation is in progress, please complete or abort it first", operation)
	}

	// Fetch remote refs.
	err = scmProvider.FetchAll()
	if err != nil {
		return seederr.Wrap(err)
	}

	// Search remote for feature branch.
	allRemoteBranches, err := scmProvider.ListRemoteBranches(options.Remote)
	if err != nil {
		return seederr.Wrap(err)
	}
	allRemoteChangeBranches, err := listRemoteChangeBranch(allRemoteBranches)
	if err != nil {
		return seederr.Wrap(err)
	}
	matchedRemoteChangeBranches := []RemoteChangeBranch{}
	for _, remoteChangeBranch := range allRemoteChangeBranches {
		if remoteChangeBranch.Remote == options.Remote && remoteChangeBranch.FeatureName == options.FeatureName {
			if options.Owner != "" && remoteChangeBranch.Owner != options.Owner {
				continue
			}
			matchedRemoteChangeBranches = append(matchedRemoteChangeBranches, remoteChangeBranch)
		}
	}
	if len(matchedRemoteChangeBranches) == 0 {
		return seederr.WrapErrorf("no remote change branch found for %v on %v", options.FeatureName, options.Remote)
	}
	if len(matchedRemoteChangeBranches) > 1 {
		return seederr.WrapErrorf("multiple remote change branches found for %v on %v: %v", options.FeatureName, options.Remote, matchedRemoteChangeBranches)
	}
	remoteChangeBranch := matchedRemoteChangeBranches[0].String()

	activeBranch, err := scmProvider.GetWorktreeBranch("")
	if err != nil {
		return seederr.Wrap(err)
	}
	if activeBranch == devBranch {
		// Must not apply on the tail of dev branch, apply to the top of dev branch instead.
		devTrackingBranch, err := scmProvider.GetBranchTracking(activeBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		// The active branch will be restored during the sync, so we don't need to worry about checking out back.
		err = scmProvider.Checkout("", devTrackingBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		activeBranch = devTrackingBranch
	}

	// Walk chain and validate, find child of activeBranch.
	currentBranch := devBranch
	childBranch := ""
	for currentBranch != ("base/" + devBranch) {
		currentTrackingBranch, err := scmProvider.GetBranchTracking(currentBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		if !strings.HasPrefix(currentTrackingBranch, "change/") && currentTrackingBranch != ("base/"+devBranch) {
			return seederr.WrapErrorf("tracking chain is broken for %v (points to %v)", currentBranch, currentTrackingBranch)
		}
		if currentTrackingBranch == activeBranch {
			childBranch = currentBranch
		}
		currentBranch = currentTrackingBranch
	}
	if childBranch == "" {
		return seederr.WrapErrorf("branch %v is not in the tracking chain of %v", activeBranch, devBranch)
	}

	changeBranch := "change/" + options.FeatureName
	_, err = scmProvider.GetBranch(changeBranch)
	if err != nil && !errors.Is(err, scm.ErrBranchNotFound) {
		return seederr.Wrap(err)
	}
	if err == nil {
		return seederr.WrapErrorf("change branch %v already exists, apply the same change in the same local repo is unsupported", changeBranch)
	}

	// Create and checkout change branch.
	startPoint, err := scmProvider.GetCommitId(activeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.CreateBranch(changeBranch, startPoint, activeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.SetBranchTracking(childBranch, changeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.Checkout("", changeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Cherry-pick commits from the remote feature branch.
	mergeBase, err := scmProvider.GetMergeBaseCommitId(activeBranch, remoteChangeBranch)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.ApplyCommitRange("", mergeBase, remoteChangeBranch)
	if err != nil {
		return seederr.WrapErrorf("cherry-pick failed; resolve the conflict and continue with 'git cherry-pick --continue', then run 'nd sync': %w", err)
	}
	seedlog.Infof("Applied %v as %v", remoteChangeBranch, changeBranch)

	// Sync the chain.
	err = NdSync(scmProvider, NdSyncOptions{})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
