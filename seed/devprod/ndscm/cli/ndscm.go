package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	_ "github.com/ndscm/theseed/seed/devprod/ndscm/scm/git"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// TODO(nagi): support subcommand flags, e.g. nd dev --foo, nd review --bar, etc.

func clientCommand(command string, args []string) error {
	cc := &clientcore.ClientCore{}
	err := cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	switch command {
	case "cut":
		err := cc.NdCut(args)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "dev":
		err := cc.NdDev(args)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "review":
		err := cc.NdReview(args)
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
	err := seedinit.Initialize(
		seedinit.WithEnvPrefix("ND_"),
		seedinit.WithSystemEnv("ndscm/ndscm.env"),
		seedinit.WithUserEnv("ndscm/ndscm.env"),
		seedinit.WithAncestorEnv("ndscm.env"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(flag.Args()) < 1 {
		fmt.Printf("ndscm is not-distributed source code manager\n")
		return nil
	}
	switch flag.Arg(0) {
	case "cut":
		err := clientCommand("cut", flag.Args()[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "dev":
		err := clientCommand("dev", flag.Args()[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "review":
		err := clientCommand("review", flag.Args()[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "setup":
		err := clientCommand("setup", flag.Args()[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "shell":
		err := clientCommand("shell", flag.Args()[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "sync":
		err := clientCommand("sync", flag.Args()[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	default:
		flag.PrintDefaults()
		return seederr.WrapErrorf("unknown command %v", flag.Arg(0))
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
