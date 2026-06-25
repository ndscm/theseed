package golinkroute

import (
	"context"
	"net/http"
	"net/http/httputil"
	"net/url"
	"os"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/authfe"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagGolinkServiceServer = seedflag.DefineString(
	"golink_service_server", "http://127.0.0.1:4656",
	"URL of Golink service server",
)
var flagGolinkOpenidDiscoveryUrl = seedflag.DefineString(
	"golink_openid_discovery_url", "",
	"Discovery URL of Golink OpenID provider. If not specified, will use default OpenID Discovery URL",
)
var flagGolinkOpenidClientId = seedflag.DefineString(
	"golink_openid_client_id", "",
	"Client ID for Golink OpenID provider",
)
var flagGolinkOpenidClientSecretFile = seedflag.DefineString(
	"golink_openid_client_secret_file", "",
	"Client secret file for Golink OpenID provider. If not specified, will use SFE OpenID client with signed assertion",
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
	clientSecretFile := flagGolinkOpenidClientSecretFile.Get()
	clientSecret := ""
	if clientSecretFile != "" {
		clientSecretBytes, err := os.ReadFile(clientSecretFile)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		clientSecret = strings.TrimSpace(string(clientSecretBytes))
	}
	golinkOpenidClient := openid.NewOpenidClient(
		discoveryUrl, clientId, clientSecret,
	)
	if clientSecret == "" && sfeOpenidClient != nil {
		refreshCtx := context.Background()
		sfeTokenSource, err := sfeOpenidClient.GetTokenSource(refreshCtx, []string{"openid"})
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		golinkOpenidClient = openid.NewOpenidClientWithAssertion(
			discoveryUrl, clientId, sfeTokenSource,
		)
	}
	provider := openid.NewOpenidProvider(golinkOpenidClient, "golink_")
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
