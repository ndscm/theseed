package main

import (
	"flag"
	"fmt"
	"os"

	"github.com/ndscm/theseed/seed/devprod/ndscm/common"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// TODO(nagi): support subcommand flags, e.g. nd dev --foo, nd review --bar, etc.

func run() error {
	err := seedinit.Initialize(
		seedinit.WithEnvPrefix("ND_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}
	ndConfig, err := common.LoadConfig()
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(flag.Args()) < 1 {
		fmt.Printf("ndscm is not-distributed source code manager\n")
		return nil
	}
	switch flag.Arg(0) {
	case "cut":
		err := NdCut(flag.Args(), ndConfig)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "dev":
		err := NdDev(flag.Args(), ndConfig)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "review":
		err := NdReview(flag.Args(), ndConfig)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "setup":
		err := NdSetup(flag.Args(), ndConfig)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "shell":
		err := NdShell(flag.Args(), ndConfig)
		if err != nil {
			return seederr.Wrap(err)
		}
	case "sync":
		err := NdSync(flag.Args(), ndConfig)
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
