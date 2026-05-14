package main

import (
	"fmt"
	"os"

	_ "github.com/ndscm/theseed/seed/devprod/ndscm/scm/git"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

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
	case "apply":
		err := ndApply(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "bootstrap":
		err := ndBootstrap(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "change":
		err := ndChange(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "connect":
		err := ndConnect(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "cut":
		err := ndCut(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "dev":
		err := ndDev(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "format":
		err := ndFormat(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "run":
		err := ndRun(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "setup":
		err := ndSetup(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "shell":
		err := ndShell(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "submit":
		err := ndSubmit(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "sync":
		err := ndSync(args[1:])
		if err != nil {
			return seederr.Wrap(err)
		}
	case "uncut":
		err := ndUncut(args[1:])
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
