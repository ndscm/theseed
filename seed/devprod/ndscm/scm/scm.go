package scm

import (
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagScm = seedflag.DefineString("scm", "ndscm", "the scm backend")

func ScmName() (string, error) {
	scm := flagScm.Get()
	switch scm {
	case "git":
		return "git", nil
	}
	return "", seederr.WrapErrorf("scm is unsupported: %v", scm)
}
