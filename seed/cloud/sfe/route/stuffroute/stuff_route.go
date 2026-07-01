package stuffroute

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

var flagStuffServiceServer = seedflag.DefineString(
	"stuff_service_server", "http://127.0.0.1:7883",
	"URL of the Stuff service server.",
)
var flagStuffOpenidDiscoveryUrl = seedflag.DefineString(
	"stuff_openid_discovery_url", "",
	"Discovery URL for the Stuff OpenID provider, used only when a client ID is set. Defaults to the standard OpenID discovery URL when left empty.",
)
var flagStuffOpenidClientId = seedflag.DefineString(
	"stuff_openid_client_id", "",
	"Client ID for the Stuff OpenID provider. Falls back to the SFE OpenID client when left empty.",
)
var flagStuffOpenidClientSecret = seedflag.DefineSecret(
	"stuff_openid_client_secret",
	"Client secret for the Stuff OpenID provider. When left empty, a signed assertion is attempted instead.",
)

type StuffRoute struct {
	authHandler *authfe.AuthHandler

	next http.Handler
}

func (p *StuffRoute) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	if strings.HasPrefix(r.URL.Path, "/auth/") {
		p.authHandler.ServeHTTP(w, r)
		return
	}
	p.next.ServeHTTP(w, r)
}

var _ http.Handler = (*StuffRoute)(nil)

func CreateStuffRoute(sfeOpenidClient *openid.OpenidClient) (*StuffRoute, error) {
	// Create Auth Handler
	discoveryUrl := flagStuffOpenidDiscoveryUrl.Get()
	if discoveryUrl == "" {
		discoveryUrl = openid.OpenidDiscoveryUrlFlag()
	}
	clientId := flagStuffOpenidClientId.Get()
	clientSecret, err := flagStuffOpenidClientSecret.LoadString()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	client, err := signedjwt.WrapOpenidClient(
		discoveryUrl, clientId, clientSecret, sfeOpenidClient,
	)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	provider := openid.NewOpenidProvider(client, "stuff_")
	authHandler := authfe.NewAuthHandler(provider)

	// Create Reverse Proxy
	serverUrl, err := url.Parse(flagStuffServiceServer.Get())
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

	// Create Stuff Route
	route := &StuffRoute{
		authHandler: authHandler,
		next:        next,
	}
	return route, nil
}
