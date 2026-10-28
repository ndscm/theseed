package main

import (
	"runtime"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndFormatFlags struct {
	all     *seedflag.BoolFlag
	changed *seedflag.BoolFlag
}

func parseNdFormatFlags(args []string) (ndFormatFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-format")
	cmdFlags := ndFormatFlags{}
	cmdFlags.all = cf.DefineBool("all", false, "Check all watchers regardless of cached modified time")
	cmdFlags.changed = cf.DefineBool("changed", true, "Also format files changed in the last commit")
	cmdArgs, err := cf.Parse(args, &seedflag.ParseOptions{
		AnywhereFlag: true,
	})
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndFormat(args []string) error {
	cmdFlags, cmdArgs, err := parseNdFormatFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	filePaths := cmdArgs
	skipScan := len(filePaths) > 0
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdRun(clientcore.NdRunOptions{
		Workers: runtime.NumCPU(),
		All:     cmdFlags.all.Get(),
		Changed: cmdFlags.changed.Get(),

		SkipScmScan: skipScan,
		CarePaths:   filePaths,

		Phases: []string{"format"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

type ndVendorFlags struct {
	all *seedflag.BoolFlag
}

func parseNdVendorFlags(args []string) (ndVendorFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-vendor")
	cmdFlags := ndVendorFlags{}
	cmdFlags.all = cf.DefineBool("all", false, "Check all watchers regardless of cached modified time")
	cmdArgs, err := cf.Parse(args, &seedflag.ParseOptions{
		AnywhereFlag: true,
	})
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndVendor(args []string) error {
	cmdFlags, cmdArgs, err := parseNdVendorFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-vendor usage: ndscm vendor")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdRun(clientcore.NdRunOptions{
		Workers: runtime.NumCPU(),
		All:     cmdFlags.all.Get(),
		Changed: false,

		Phases: []string{"vendor"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

type ndBootstrapFlags struct {
	all *seedflag.BoolFlag
}

func parseNdBootstrapFlags(args []string) (ndBootstrapFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-bootstrap")
	cmdFlags := ndBootstrapFlags{}
	cmdFlags.all = cf.DefineBool("all", false, "Check all watchers regardless of cached modified time")
	cmdArgs, err := cf.Parse(args, &seedflag.ParseOptions{
		AnywhereFlag: true,
	})
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndBootstrap(args []string) error {
	cmdFlags, cmdArgs, err := parseNdBootstrapFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-bootstrap usage: ndscm bootstrap")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdRun(clientcore.NdRunOptions{
		Workers: runtime.NumCPU(),
		All:     cmdFlags.all.Get(),
		Changed: false,

		Phases: []string{"vendor", "bootstrap"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

type ndTidyFlags struct {
	all *seedflag.BoolFlag
}

func parseNdTidyFlags(args []string) (ndTidyFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-tidy")
	cmdFlags := ndTidyFlags{}
	cmdFlags.all = cf.DefineBool("all", false, "Check all watchers regardless of cached modified time")
	cmdArgs, err := cf.Parse(args, &seedflag.ParseOptions{
		AnywhereFlag: true,
	})
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndTidy(args []string) error {
	cmdFlags, cmdArgs, err := parseNdTidyFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(cmdArgs) != 0 {
		return seederr.WrapErrorf("nd-tidy usage: ndscm tidy")
	}
	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdRun(clientcore.NdRunOptions{
		Workers: runtime.NumCPU(),
		All:     cmdFlags.all.Get(),
		Changed: false,

		Phases: []string{"vendor", "bootstrap", "tidy"},
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
