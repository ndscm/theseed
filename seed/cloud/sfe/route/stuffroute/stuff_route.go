package stuffroute

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"

	"github.com/ndscm/theseed/seed/infra/auth/go/authfe"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagStuffServiceServer = seedflag.DefineString("stuff_service_server", "", "URL of Stuff service server")
var flagStuffOpenidDiscoveryUrl = seedflag.DefineString("stuff_openid_discovery_url", "http://127.0.0.1:8080/realms/ndscm/.well-known/openid-configuration", "Discovery URL of Stuff OpenID provider")
var flagStuffOpenidClientId = seedflag.DefineString("stuff_openid_client_id", "", "Client ID for Stuff OpenID provider")
var flagStuffOpenidClientSecret = seedflag.DefineSecret(
	"stuff_openid_client_secret",
	"Client secret for Stuff OpenID provider",
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

func CreateStuffRoute(transport http.RoundTripper) (*StuffRoute, error) {
	discoveryUrl := flagStuffOpenidDiscoveryUrl.Get()
	clientId := flagStuffOpenidClientId.Get()
	clientSecret, err := flagStuffOpenidClientSecret.LoadString()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	provider := openid.NewOpenidProvider(
		openid.NewOpenidClient(discoveryUrl, clientId, clientSecret), "stuff_")
	authHandler := authfe.NewAuthHandler(provider)
	serverUrl, err := url.Parse(flagStuffServiceServer.Get())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	reverseProxy := &httputil.ReverseProxy{
		Transport: transport,
		Rewrite: func(r *httputil.ProxyRequest) {
			r.SetURL(serverUrl)
			r.SetXForwarded()
		},
	}
	next := authfe.InterceptSessionAuthorizationMiddleware(reverseProxy, provider)
	route := &StuffRoute{
		authHandler: authHandler,
		next:        next,
	}
	return route, nil
}
