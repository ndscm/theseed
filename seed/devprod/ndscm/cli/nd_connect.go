package main

import (
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndConnectFlags struct {
	noDev *seedflag.BoolFlag
}

func parseNdConnectFlags(args []string) (ndConnectFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-connect")
	cmdFlags := ndConnectFlags{}
	cmdFlags.noDev = cf.DefineBool("no-dev", false, "Skip creating a dev worktree after connecting")
	cmdArgs, err := cf.Parse(args,
		seedflag.WithAnywhereFlag(true),
	)
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndConnect(args []string) error {
	cmdFlags, cmdArgs, err := parseNdConnectFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 2 {
		return seederr.WrapErrorf("nd-connect usage: nd connect <repo-identifier> <repo-endpoint>")
	}
	repoIdentifier := strings.TrimSpace(cmdArgs[0])
	repoEndpoint := strings.TrimSpace(cmdArgs[1])
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdConnect(clientcore.NdConnectOptions{
		NoDev:          cmdFlags.noDev.Get(),
		RepoIdentifier: repoIdentifier,
		RepoEndpoint:   repoEndpoint,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
