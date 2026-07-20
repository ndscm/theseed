package clientcore

import (
	"fmt"
	"os"
	"path/filepath"
	"regexp"
	"slices"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/devprod/ndscm/user"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdSecretOptions struct {
	Space string
	User  bool

	Args []string
}

func getSecretWorktree(
	monorepoHome string, userHandle string, space string,
) (string, string, bool) {
	worktreeName := "secret/" + space
	if userHandle != "" {
		worktreeName = userHandle + "/secret/" + space
	}
	worktreePath := filepath.Join(monorepoHome, worktreeName)

	exists := false
	worktreeStat, err := os.Stat(worktreePath)
	if err == nil && worktreeStat.IsDir() {
		exists = true
	}

	return worktreeName, worktreePath, exists
}

func createSecretWorktree(
	scmProvider scm.Provider,
	monorepoHome string, userHandle string, space string,
) (string, error) {
	worktreeName, worktreePath, exists := getSecretWorktree(monorepoHome, userHandle, space)
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
		// A secret space always starts as an orphan branch with no shared history.
		err = scmProvider.CreateOrphanBranch(branchName, "secret: init")
		if err != nil {
			return "", seederr.WrapErrorf("failed to create orphan branch %v: %v", branchName, err)
		}
		err = scmProvider.PushBranch(branchName, remote, remoteBranchName)
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

var secretSpaceRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func NdSecret(scmProvider scm.Provider, options NdSecretOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-secret should not run with --shell-eval")
	}
	getPath := false
	secretPath := ""
	if len(options.Args) > 0 {
		subcommand := options.Args[0]
		switch subcommand {
		case "get-path":
			if len(options.Args) != 2 {
				return seederr.WrapErrorf("nd-secret usage: nd secret [--space=<space>] [--user] get-path <secret-path>")
			}
			getPath = true
			secretPath = strings.TrimSpace(options.Args[1])
		default:
			return seederr.WrapErrorf("unknown nd-secret subcommand %v", subcommand)
		}
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
	if !secretSpaceRegex.MatchString(space) {
		return seederr.WrapErrorf("only letters, digits, - are allowed for space")
	}
	userHandle := ""
	if options.User {
		userHandle, err = user.CurrentUserHandle()
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	_, worktreePath, exists := getSecretWorktree(monorepoHome, userHandle, space)
	if !exists {
		worktreePath, err = createSecretWorktree(scmProvider, monorepoHome, userHandle, space)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if getPath {
		if filepath.IsAbs(secretPath) {
			return seederr.WrapErrorf("secret path must be relative: %v", secretPath)
		}
		secretAbsPath := filepath.Join(worktreePath, secretPath)
		secretRelPath, err := filepath.Rel(worktreePath, secretAbsPath)
		if err != nil {
			return seederr.Wrap(err)
		}
		if strings.HasPrefix(secretRelPath, "..") {
			return seederr.WrapErrorf("secret path escapes the secret worktree: %v", secretPath)
		}
		fmt.Printf("%v\n", secretAbsPath)
	}
	return nil
}
