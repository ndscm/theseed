package openid

import (
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagOpenidDiscoveryUrl = seedflag.DefineString("openid_discovery_url", "", "Discovery URL of OpenID provider, usually ends with /.well-known/openid-configuration")

func OpenidDiscoveryUrlFlag() string {
	return flagOpenidDiscoveryUrl.Get()
}
