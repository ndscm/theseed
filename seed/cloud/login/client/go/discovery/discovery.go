package discovery

import (
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagLoginOpenidDiscoveryUrl = seedflag.DefineString("login_openid_discovery_url", "http://127.0.0.1:8080/realms/ndscm/.well-known/openid-configuration", "Discovery URL of Login OpenID provider")

func LoginOpenidDiscoveryUrl() string {
	return flagLoginOpenidDiscoveryUrl.Get()
}
