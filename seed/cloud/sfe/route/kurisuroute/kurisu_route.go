package kurisuroute

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

var flagKurisuServiceServer = seedflag.DefineString("kurisu_service_server", "", "URL of Kurisu service server")
var flagKurisuOpenidDiscoveryUrl = seedflag.DefineString("kurisu_openid_discovery_url", "http://127.0.0.1:8080/realms/ndscm/.well-known/openid-configuration", "Discovery URL of Kurisu OpenID provider")
var flagKurisuOpenidClientId = seedflag.DefineString("kurisu_openid_client_id", "", "Client ID for Kurisu OpenID provider")
var flagKurisuOpenidClientSecretFile = seedflag.DefineString("kurisu_openid_client_secret_file", "", "Client secret file for Kurisu OpenID provider")

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

func CreateKurisuRoute(transport http.RoundTripper) (*KurisuRoute, error) {
	discoveryUrl := flagKurisuOpenidDiscoveryUrl.Get()
	clientId := flagKurisuOpenidClientId.Get()
	clientSecretFile := flagKurisuOpenidClientSecretFile.Get()
	clientSecret := ""
	if clientSecretFile != "" {
		clientSecretBytes, err := os.ReadFile(clientSecretFile)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		clientSecret = strings.TrimSpace(string(clientSecretBytes))
	}
	provider := loginopenid.NewUserOpenidProvider(
		clientopenid.NewOpenidProvider(discoveryUrl, clientId, clientSecret), "kurisu_")
	authHandler := authfe.NewAuthHandler(provider)
	serverUrl, err := url.Parse(flagKurisuServiceServer.Get())
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
	route := &KurisuRoute{
		authHandler: authHandler,
		next:        next,
	}
	return route, nil
}
