package kurisuroute

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/cloud/sfe/signedjwt"
	"github.com/ndscm/theseed/seed/infra/auth/go/authfe"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagKurisuServiceServer = seedflag.DefineString(
	"kurisu_service_server", "",
	"URL of the Kurisu service server.",
)
var flagKurisuOpenidDiscoveryUrl = seedflag.DefineString(
	"kurisu_openid_discovery_url", "",
	"Discovery URL for the Kurisu OpenID provider, used only when a client ID is set. Defaults to the standard OpenID discovery URL when left empty.",
)
var flagKurisuOpenidClientId = seedflag.DefineString(
	"kurisu_openid_client_id", "",
	"Client ID for the Kurisu OpenID provider. Falls back to the SFE OpenID client when left empty.",
)
var flagKurisuOpenidClientSecret = seedflag.DefineSecret(
	"kurisu_openid_client_secret",
	"Client secret for the Kurisu OpenID provider. When left empty, a signed assertion is attempted instead.",
)

type KurisuRoute struct {
	authHandler *authfe.AuthHandler

	next http.Handler
}

func (p *KurisuRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/auth/") {
		p.authHandler.ServeHTTP(w, r)
		return
	}
	p.next.ServeHTTP(w, r)
}

var _ http.Handler = (*KurisuRoute)(nil)

func CreateKurisuRoute(sfeOpenidClient *openid.OpenidClient) (*KurisuRoute, error) {
	// Create Auth Handler
	discoveryUrl := flagKurisuOpenidDiscoveryUrl.Get()
	if discoveryUrl == "" {
		discoveryUrl = openid.OpenidDiscoveryUrlFlag()
	}
	clientId := flagKurisuOpenidClientId.Get()
	clientSecret, err := flagKurisuOpenidClientSecret.LoadString()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	client, err := signedjwt.WrapOpenidClient(
		discoveryUrl, clientId, clientSecret, sfeOpenidClient,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	provider := openid.NewOpenidProvider(client, "kurisu_")
	authHandler := authfe.NewAuthHandler(provider)

	// Create Reverse Proxy
	serverUrl, err := url.Parse(flagKurisuServiceServer.Get())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	reverseProxy := &httputil.ReverseProxy{
		Transport: &http.Transport{
			Proxy:               nil,
			MaxIdleConns:        200,
			MaxIdleConnsPerHost: 100,
			IdleConnTimeout:     10 * time.Second,
		},
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(serverUrl)
			r.SetXForwarded()
		},
	}
	next := authfe.InterceptSessionAuthorizationMiddleware(reverseProxy, provider)

	// Create Kurisu Route
	route := &KurisuRoute{
		authHandler: authHandler,
		next:        next,
	}
	return route, nil
}
