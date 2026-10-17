package runner

import (
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagBuildSystems = seedflag.DefineStringList(
	"build_system", []string{"bazel"},
	"The monorepo build system, can be repeated as candidates")

func BuildSystems() []string {
	return flagBuildSystems.Get()
}
