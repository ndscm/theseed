package main

import (
	"fmt"
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	_ "github.com/ndscm/theseed/seed/devprod/ndscm/scm/git"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// TODO(nagi): support subcommand flags, e.g. nd dev --foo, nd submit --bar, etc.

func clientCommand(command string, args []string) error {
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	switch command {
	case "dev":
		err := cc.NdDev(args)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "setup":
		err := cc.NdSetup(args)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "shell":
		err := cc.NdShell(args)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "sync":
		err := cc.NdSync(args)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

func run() error {
	args, err := seedinit.Initialize(
		seedinit.WithAnywhereFlag(true),
		seedinit.WithUnknownFlag(true),
		seedinit.WithEnvPrefix("ND_"),
		seedinit.WithSystemEnv("ndscm/ndscm.env"),
		seedinit.WithUserEnv("ndscm/ndscm.env"),
		seedinit.WithAncestorEnv("ndscm.env"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(args) < 1 {
		fmt.Printf("ndscm is not-distributed source code manager\n")
		return nil
	}
	command := args[0]
	switch command {
	case "cut":
		err := ndCut(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "dev":
		err := clientCommand("dev", args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "setup":
		err := clientCommand("setup", args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "shell":
		err := clientCommand("shell", args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "submit":
		err := ndSubmit(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "sync":
		err := clientCommand("sync", args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	default:
		return seederr.WrapErrorf("unknown command %v", command)
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
