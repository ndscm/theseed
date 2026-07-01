package golinkroute

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

var flagGolinkServiceServer = seedflag.DefineString(
	"golink_service_server", "http://127.0.0.1:4656",
	"URL of the Golink service server.",
)
var flagGolinkOpenidDiscoveryUrl = seedflag.DefineString(
	"golink_openid_discovery_url", "",
	"Discovery URL for the Golink OpenID provider, used only when a client ID is set. Defaults to the standard OpenID discovery URL when left empty.",
)
var flagGolinkOpenidClientId = seedflag.DefineString(
	"golink_openid_client_id", "",
	"Client ID for the Golink OpenID provider. Falls back to the SFE OpenID client when left empty.",
)
var flagGolinkOpenidClientSecret = seedflag.DefineSecret(
	"golink_openid_client_secret",
	"Client secret for the Golink OpenID provider. When left empty, a signed assertion is attempted instead.",
)

type GolinkRoute struct {
	authHandler *authfe.AuthHandler

	next http.Handler
}

func (p *GolinkRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/auth/") {
		p.authHandler.ServeHTTP(w, r)
		return
	}
	p.next.ServeHTTP(w, r)
}

var _ http.Handler = (*GolinkRoute)(nil)

func CreateGolinkRoute(sfeOpenidClient *openid.OpenidClient) (*GolinkRoute, error) {
	// Create Auth Handler
	discoveryUrl := flagGolinkOpenidDiscoveryUrl.Get()
	if discoveryUrl == "" {
		discoveryUrl = openid.OpenidDiscoveryUrlFlag()
	}
	clientId := flagGolinkOpenidClientId.Get()
	clientSecret, err := flagGolinkOpenidClientSecret.LoadString()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	client, err := signedjwt.WrapOpenidClient(
		discoveryUrl, clientId, clientSecret, sfeOpenidClient,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	provider := openid.NewOpenidProvider(client, "golink_")
	authHandler := authfe.NewAuthHandler(provider)

	// Create Reverse Proxy
	serverUrl, err := url.Parse(flagGolinkServiceServer.Get())
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

	// Create Golink Route
	route := &GolinkRoute{
		authHandler: authHandler,
		next:        next,
	}
	return route, nil
}
