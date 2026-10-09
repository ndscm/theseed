// Package stuffroute is the multitenant counterpart of
// seed/cloud/sfe/route/stuffroute. It serves the same Stuff frontend, but
// resolves the OpenID provider per request from the supplied resolver instead
// of binding a single provider at startup, so one instance serves every tenant.
package stuffroute

import (
	"net/http"
	"net/http/httputil"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/auth/go/authfe"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
)

var flagStuffServiceServer = seedflag.DefineString(
	"stuff_service_server", "http://127.0.0.1:7883",
	"URL of the Stuff service server.",
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

var contentTypeRegexp = regexp.MustCompile("^application/(json|grpc)")

// CreateStuffRoute builds a multitenant Stuff route. resolver resolves the
// tenant's OpenID provider per request. The returned route does not establish
// the session itself; the caller is expected to wrap it with its session
// middleware, just as the sfe server does.
func CreateStuffRoute(resolver authfe.OpenidProviderResolver) (*StuffRoute, error) {
	// Create Auth Handler
	authHandler := authfe.NewMultiTenantAuthHandler(resolver, nil)

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
	next := authfe.InterceptAuthMiddleware(
		reverseProxy,
		authHandler,
		&authfe.AuthMiddlewareOptions{
			Matcher: func(r *http.Request) bool {
				return contentTypeRegexp.MatchString(r.Header.Get("Content-Type"))
			},
		},
	)

	// Create Stuff Route
	route := &StuffRoute{
		authHandler: authHandler,
		next:        next,
	}
	return route, nil
}
