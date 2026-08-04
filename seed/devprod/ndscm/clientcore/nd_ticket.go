package clientcore

import (
	"os"
	"path/filepath"
	"regexp"
	"slices"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdTicketOptions struct {
	Space string

	Args []string
}

func getTicketWorktree(
	monorepoHome string, space string,
) (string, string, bool) {
	worktreeName := "ticket/" + space
	worktreePath := filepath.Join(monorepoHome, worktreeName)

	exists := false
	worktreeStat, err := os.Stat(worktreePath)
	if err == nil && worktreeStat.IsDir() {
		exists = true
	}

	return worktreeName, worktreePath, exists
}

func createTicketWorktree(
	scmProvider scm.Provider,
	monorepoHome string, space string,
) (string, error) {
	worktreeName, worktreePath, exists := getTicketWorktree(monorepoHome, space)
	if exists {
		return "", seederr.WrapErrorf("worktree path already exists. path=%v", worktreePath)
	}
	branchName := worktreeName
	remote := "origin"
	remoteBranchName := branchName
	remoteTracking := remote + "/" + remoteBranchName
	err := scmProvider.FetchAll()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	remoteBranches, err := scmProvider.ListRemoteBranches(remote)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	if slices.Contains(remoteBranches, remoteTracking) {
		err = scmProvider.CreateBranch(branchName, remoteTracking, remoteTracking)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create branch %v: %v", branchName, err)
		}
	} else {
		// A ticket space always starts as an orphan branch with no shared history.
		message := "ticket: init"
		err = scmProvider.CreateOrphanBranch(branchName, message)
		if err != nil {
			return "", seederr.WrapErrorf("failed to create orphan branch %v: %v", branchName, err)
		}
		err = scmProvider.PushBranch(branchName, remote, remoteBranchName, true)
		if err != nil {
			return "", seederr.WrapErrorf("failed to push branch %v to %v: %v", branchName, remote, err)
		}
		err = scmProvider.SetBranchTracking(branchName, remoteTracking)
		if err != nil {
			return "", seederr.WrapErrorf("failed to set tracking for branch %v: %v", branchName, err)
		}
	}
	newWorktreePath, err := scmProvider.CreateWorktree(monorepoHome, branchName)
	if err != nil {
		return "", seederr.WrapErrorf("failed to create worktree for branch %v: %v", branchName, err)
	}
	if newWorktreePath != worktreePath {
		return "", seederr.WrapErrorf("unexpected new worktree path: %v (expected: %v)", newWorktreePath, worktreePath)
	}
	return newWorktreePath, nil
}

func NdTicketSync(
	scmProvider scm.Provider,
	monorepoHome string, space string,
) error {
	worktreeName, worktreePath, exists := getTicketWorktree(monorepoHome, space)
	if !exists {
		newWorktreePath, err := createTicketWorktree(scmProvider, monorepoHome, space)
		if err != nil {
			return seederr.Wrap(err)
		}
		worktreePath = newWorktreePath
	}

	// The ticket worktree must have its own branch checked out; refuse to commit
	// into it if some other branch has been swapped in.
	branchName, err := scmProvider.GetWorktreeBranch(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	if branchName != worktreeName {
		return seederr.WrapErrorf("ticket worktree has unexpected branch checked out: %v (expected: %v)", branchName, worktreeName)
	}

	err = scmProvider.PullRebase(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.PushBranch(branchName, "origin", branchName, false)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

var ticketSpaceRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func NdTicket(scmProvider scm.Provider, options NdTicketOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-ticket should not run with --shell-eval")
	}
	subcommand := "sync"
	if len(options.Args) > 0 {
		subcommand = options.Args[0]
	}
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = scmProvider.QuickVerifyMonorepo()
	if err != nil {
		return seederr.Wrap(err)
	}
	space := options.Space
	if space == "" {
		space = "main"
	}
	if !ticketSpaceRegex.MatchString(space) {
		return seederr.WrapErrorf("only letters, digits, - are allowed for space")
	}
	switch subcommand {
	case "sync":
		err := NdTicketSync(scmProvider, monorepoHome, space)
		if err != nil {
			return seederr.Wrap(err)
		}
	default:
		return seederr.WrapErrorf("unknown nd-ticket subcommand %v", subcommand)
	}
	return nil
}
