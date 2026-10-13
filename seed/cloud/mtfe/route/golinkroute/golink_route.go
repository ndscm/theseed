package golinkroute

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/authfe"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagGolinkServiceServer = seedflag.DefineString(
	"golink_service_server", "http://127.0.0.1:4656",
	"URL of the Golink service server.",
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

func CreateGolinkRoute(provider authfe.DynamicOpenidProvider) (*GolinkRoute, error) {
	// Create Auth Handler
	authHandler := authfe.NewDynamicAuthHandler(provider, nil)

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

	// Golink server needs authentication for http redirection
	next := authfe.InterceptAuthMiddleware(reverseProxy, authHandler, nil)

	// Create Golink Route
	route := &GolinkRoute{
		authHandler: authHandler,
		next:        next,
	}
	return route, nil
}
