package main

import (
	"fmt"
	"os"
	"path/filepath"

	"github.com/ndscm/theseed/seed/devprod/rbe/bes/go/bep"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func run() error {
	args, err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	if len(args) > 1 {
		return seederr.WrapErrorf("usage: printbep [file.bep]")
	}

	bepFilePath := ".bep"
	bazelBuildWorkspaceDirectory := os.Getenv("BUILD_WORKSPACE_DIRECTORY")
	if bazelBuildWorkspaceDirectory != "" {
		bepFilePath = filepath.Join(bazelBuildWorkspaceDirectory, ".bep")
	}
	if len(args) == 1 {
		bepFilePath = args[0]
	}

	data, err := os.ReadFile(bepFilePath)
	if err != nil {
		return seederr.Wrap(err)
	}
	events, err := bep.ParseBuildEventProtos(data)
	if err != nil {
		return seederr.Wrap(err)
	}
	out, err := events.DumpJson(true)
	if err != nil {
		return seederr.Wrap(err)
	}
	fmt.Printf("%s\n", string(out))
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
