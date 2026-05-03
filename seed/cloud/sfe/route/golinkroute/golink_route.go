package golinkroute

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"

	"github.com/ndscm/theseed/seed/infra/auth/go/authfe"
	"github.com/ndscm/theseed/seed/infra/auth/go/clientopenid"
	"github.com/ndscm/theseed/seed/infra/auth/go/loginopenid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagGolinkServiceServer = seedflag.DefineString("golink_service_server", "", "URL of Golink service server")
var flagGolinkOpenidDiscovery = seedflag.DefineString("golink_openid_discovery", "http://127.0.0.1:8080/realms/ndscm/.well-known/openid-configuration", "Discovery URL of Golink OpenID provider")
var flagGolinkOpenidClientId = seedflag.DefineString("golink_openid_client_id", "", "Client ID for Golink OpenID provider")
var flagGolinkOpenidClientSecretFile = seedflag.DefineString("golink_openid_client_secret_file", "", "Client secret file for Golink OpenID provider")

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

func CreateGolinkRoute(transport http.RoundTripper) (*GolinkRoute, error) {
	discovery := flagGolinkOpenidDiscovery.Get()
	clientId := flagGolinkOpenidClientId.Get()
	clientSecretFile := flagGolinkOpenidClientSecretFile.Get()
	clientSecret := ""
	if clientSecretFile != "" {
		clientSecretBytes, err := os.ReadFile(clientSecretFile)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		clientSecret = strings.TrimSpace(string(clientSecretBytes))
	}
	provider := loginopenid.NewUserOpenidProvider(
		clientopenid.NewOpenidProvider(discovery, clientId, clientSecret), "golink_")
	authHandler := authfe.NewAuthHandler(provider)
	serverUrl, err := url.Parse(flagGolinkServiceServer.Get())
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
	route := &GolinkRoute{
		authHandler: authHandler,
		next:        next,
	}
	return route, nil
}
