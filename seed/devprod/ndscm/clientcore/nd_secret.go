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
		message := ""
		if userHandle != "" {
			message = userHandle + ": "
		}
		message += "secret: init"
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

// bfsChangedFiles returns the changed file paths in breadth-first order from
// the worktree root: shallower paths first, ties broken lexicographically.
// Sorting by (depth, path) is equivalent to a level-order walk of the file
// tree with lexicographically ordered children, because paths sharing a parent
// share a prefix and therefore stay grouped within each depth.
func bfsChangedFiles(paths []string) []string {
	ordered := slices.Clone(paths)
	slices.SortFunc(ordered, func(a, b string) int {
		depthA := strings.Count(a, "/")
		depthB := strings.Count(b, "/")
		if depthA != depthB {
			return depthA - depthB
		}
		return strings.Compare(a, b)
	})
	return ordered
}

func NdSecretSync(
	scmProvider scm.Provider,
	monorepoHome string, userHandle string, space string,
) error {
	worktreeName, worktreePath, exists := getSecretWorktree(monorepoHome, userHandle, space)
	if !exists {
		newWorktreePath, err := createSecretWorktree(scmProvider, monorepoHome, userHandle, space)
		if err != nil {
			return seederr.Wrap(err)
		}
		worktreePath = newWorktreePath
	}

	// The secret worktree must have its own branch checked out; refuse to commit
	// into it if some other branch has been swapped in.
	branchName, err := scmProvider.GetWorktreeBranch(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	if branchName != worktreeName {
		return seederr.WrapErrorf("secret worktree has unexpected branch checked out: %v (expected: %v)", branchName, worktreeName)
	}

	// Move the entire worktree out of the staging area, so every secret change
	// starts unstaged and is committed one file at a time below.
	err = scmProvider.UpdateStagingArea(worktreePath, ".", false)
	if err != nil {
		return seederr.Wrap(err)
	}

	dirtyFiles, err := scmProvider.ListDirtyFiles(worktreePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	changedFiles := map[string]string{}
	for _, f := range dirtyFiles {
		changedFiles[f.To] = f.Status
	}

	messagePrefix := ""
	if userHandle != "" {
		messagePrefix += userHandle + ": "
	}
	messagePrefix += "secret: "

	paths := make([]string, 0, len(changedFiles))
	for path := range changedFiles {
		paths = append(paths, path)
	}
	for _, path := range bfsChangedFiles(paths) {
		// The git porcelain worktree status determines the commit verb.
		action := "update"
		status := changedFiles[path]
		if status == "??" {
			action = "create"
		} else if len(status) >= 2 && status[1] == 'D' {
			action = "remove"
		}
		message := messagePrefix + action + " " + path
		err := scmProvider.UpdateStagingArea(worktreePath, path, true)
		if err != nil {
			return seederr.Wrap(err)
		}
		err = scmProvider.CreateCommit(worktreePath, message)
		if err != nil {
			return seederr.Wrap(err)
		}
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

func NdSecretGetPath(
	monorepoHome string, userHandle string, space string,
	args []string,
) error {
	if len(args) != 1 {
		return seederr.WrapErrorf("nd-secret usage: nd secret [--space=<space>] [--user] get-path <secret-path>")
	}
	secretPath := strings.TrimSpace(args[0])
	_, worktreePath, exists := getSecretWorktree(monorepoHome, userHandle, space)
	if !exists {
		syncCommand := "nd secret"
		if userHandle != "" {
			syncCommand += " --user"
		}
		if space != "" && space != "main" {
			syncCommand += " --space \"" + space + "\""
		}
		syncCommand += " sync"
		return seederr.WrapErrorf("secret worktree is not initialized, run `%s` first.", syncCommand)
	}
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
	return nil
}

var secretSpaceRegex = regexp.MustCompile(`^[a-z0-9-]+$`)

func NdSecret(scmProvider scm.Provider, options NdSecretOptions) error {
	if seedshell.ShellEval() {
		return seederr.WrapErrorf("nd-secret should not run with --shell-eval")
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
	switch subcommand {
	case "sync":
		err := NdSecretSync(scmProvider, monorepoHome, userHandle, space)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "get-path":
		err := NdSecretGetPath(monorepoHome, userHandle, space, options.Args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	default:
		return seederr.WrapErrorf("unknown nd-secret subcommand %v", subcommand)
	}
	return nil
}
