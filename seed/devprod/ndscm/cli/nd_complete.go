package main

import (
	"strconv"
	"strings"

	"github.com/ndscm/theseed/seed/devprod/ndscm/clientcore"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

type ndCompleteFlags struct {
	cur   *seedflag.StringFlag
	cword *seedflag.StringFlag
}

func parseNdCompleteFlags(args []string) (ndCompleteFlags, []string, error) {
	cf := seedflag.NewCommandFlags("nd-complete")
	cmdFlags := ndCompleteFlags{}
	cmdFlags.cur = cf.DefineString("cur", "", "The word currently being completed")
	cmdFlags.cword = cf.DefineString("cword", "", "The index of the word being completed")
	cmdArgs, err := cf.Parse(args, &seedflag.ParseOptions{})
	if err != nil {
		return cmdFlags, nil, seederr.Wrap(err)
	}
	seedflag.Finalize(cmdArgs)
	return cmdFlags, cmdArgs, nil
}

func ndComplete(args []string) error {
	cmdFlags, cmdArgs, err := parseNdCompleteFlags(args)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Drop the "--" terminator that Parse preserves; what remains is the
	// forwarded COMP_WORDS, which cword indexes into directly.
	words := []string{}
	if len(cmdArgs) > 0 && cmdArgs[0] == "--" {
		words = cmdArgs[1:]
	}

	cword, err := strconv.Atoi(strings.TrimSpace(cmdFlags.cword.Get()))
	if err != nil {
		cword = len(words) - 1
	}

	cc := &clientcore.ClientCore{}
	err = cc.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	err = cc.NdComplete(clientcore.NdCompleteOptions{
		Cur:   cmdFlags.cur.Get(),
		Cword: cword,
		Words: words,
	})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
