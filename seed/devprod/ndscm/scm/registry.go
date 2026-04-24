package scm

import (
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

// The default SCM provider is "ndscm", which is used to force the user to set git explicitly.
var flagScm = seedflag.DefineString("scm", "ndscm", "The SCM provider to use (e.g. git)")

func ScmName() (string, error) {
	scm := flagScm.Get()
	switch scm {
	case "git":
		return "git", nil
	}
	return "", seederr.WrapErrorf("scm is unsupported: %v", scm)
}
