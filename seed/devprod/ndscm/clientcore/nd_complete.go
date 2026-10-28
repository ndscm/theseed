package clientcore

import (
	"fmt"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// ndCompleteCommands is the curated, user-facing set of subcommands offered as
// completions. It is intentionally narrower than the full set the CLI dispatches.
var ndCompleteCommands = []string{
	"apply",
	"bootstrap",
	"build",
	"change",
	"check",
	"connect",
	"cut",
	"dev",
	"format",
	"hook",
	"lock",
	"melt",
	"run",
	"setup",
	"submit",
	"sync",
	"test",
	"tidy",
	"uncut",
	"vendor",
}

type NdCompleteOptions struct {
	Cur string

	Cword int

	Words []string
}

// listWorktreeChanges returns the feature names of the change branches in the
// current dev worktree's tracking chain, walking dev -> change... -> base via
// GetBranchTracking the same way NdSync does. Returns an empty list when the
// workspace is not a dev worktree.
func listWorktreeChanges(scmProvider scm.Provider) ([]string, error) {
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	repo := &scm.WorkingRepo{MonorepoHome: monorepoHome}
	worktreeName, _, err := scmProvider.GetCurrentWorktree(repo)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if !scm.IsBranchType(worktreeName, "dev") {
		return []string{}, nil
	}
	baseBranch, err := scm.GetBaseBranchName(worktreeName)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// Change branches derived from this dev branch all share its owner and
	// focus: "<owner>/change/<focus>/<featureName>". Building the prefix with an
	// empty feature name yields "<owner>/change/<focus>/" to strip back off.
	changePrefix, err := scm.GetChangeBranchName(worktreeName, "")
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	features := []string{}
	for iter := worktreeName; iter != baseBranch; {
		inspectBranch, err := scmProvider.GetBranchTracking(iter)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if inspectBranch == "" {
			return nil, seederr.WrapErrorf("tracking upstream is missing for %v", iter)
		}
		if !scm.IsBranchType(inspectBranch, "change") && inspectBranch != baseBranch {
			return nil, seederr.WrapErrorf("tracking chain is broken for %v (point to %v)", iter, inspectBranch)
		}
		if strings.HasPrefix(inspectBranch, changePrefix) {
			features = append(features, strings.TrimPrefix(inspectBranch, changePrefix))
		}
		iter = inspectBranch
	}
	return features, nil
}

// ndCompleteCandidates returns the candidates for the word at Cword, before
// prefix filtering. Words mirrors bash's COMP_WORDS: Words[0] is the program
// name, Cword 1 is the subcommand.
func ndCompleteCandidates(scmProvider scm.Provider, options NdCompleteOptions) []string {
	if options.Cword <= 1 || len(options.Words) < 2 {
		return ndCompleteCommands
	}
	candidates := []string{}
	switch options.Words[1] {
	case "change", "submit":
		// Feature-name argument comes from the change branches in the tracking
		// chain; stay quiet on error (e.g. outside a dev worktree).
		features, err := listWorktreeChanges(scmProvider)
		if err != nil {
			break
		}
		candidates = features
	}
	return candidates
}

// NdComplete prints the completion candidates for the word at Cword, one per
// line, filtered to those starting with Cur.
func NdComplete(scmProvider scm.Provider, options NdCompleteOptions) error {
	for _, candidate := range ndCompleteCandidates(scmProvider, options) {
		if !strings.HasPrefix(candidate, options.Cur) {
			continue
		}
		fmt.Println(candidate)
	}
	return nil
}
