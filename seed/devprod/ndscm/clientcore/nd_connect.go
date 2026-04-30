package clientcore

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/ndscm/scm"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/shell/go/seedshell"
)

type NdConnectOptions struct {
	NoDev bool

	RepoIdentifier string
	RepoEndpoint   string
}

func NdConnect(scmProvider scm.Provider, options NdConnectOptions) error {
	if !seedshell.ShellEval() {
		seedlog.Warnf("It's recommended to run nd-connect with --shell-eval")
	}
	monorepoHome, err := scm.MonorepoHome()
	if err != nil {
		// Ignore error
	}
	if monorepoHome == "" {
		reposHome, err := scm.ReposHome()
		if err != nil {
			return seederr.Wrap(err)
		}
		if reposHome == "" {
			return seederr.WrapErrorf("repos home must be set for nd connect")
		}
		monorepoHome = filepath.Join(reposHome, options.RepoIdentifier)
	}
	err = os.Mkdir(monorepoHome, 0755)
	if err != nil {
		if os.IsExist(err) {
			return seederr.WrapErrorf("monorepo home %v already exists", monorepoHome)
		}
		return seederr.WrapErrorf("failed to create new monorepo home %v: %w", monorepoHome, err)
	}
	remoteMainBranch, mainWorktree, err := scmProvider.Connect(options.RepoIdentifier, monorepoHome, options.RepoEndpoint)
	if err != nil {
		removeErr := os.RemoveAll(monorepoHome)
		if removeErr != nil {
			seedlog.Errorf("failed to remove monorepo home after failed connect: %v", removeErr)
		}
		return seederr.Wrap(err)
	}
	worktreePath := mainWorktree
	if !options.NoDev {
		devWorktree, err := scmProvider.CreateDevWorktree(monorepoHome, "", remoteMainBranch)
		if err != nil {
			return seederr.Wrap(err)
		}
		worktreePath = devWorktree
	}
	seedlog.Infof("Connected to %v: %v", options.RepoEndpoint, monorepoHome)

	shellEval := fmt.Sprintf("\ncd \"%v\"\n", worktreePath)
	if seedshell.Dry() {
		seedlog.Infof("Shell eval: %v", shellEval)
	}
	if seedshell.ShellEval() {
		if !seedshell.Dry() {
			fmt.Printf("%v", shellEval)
		}
	}
	return nil
}
