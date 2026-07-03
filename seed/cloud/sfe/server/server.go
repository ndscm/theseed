package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net/http"
	"os"
	"time"

	"github.com/ndscm/theseed/seed/cloud/sfe/certstore"
	"github.com/ndscm/theseed/seed/cloud/sfe/escalate"
	"github.com/ndscm/theseed/seed/cloud/sfe/route/golinkroute"
	"github.com/ndscm/theseed/seed/cloud/sfe/route/kurisuroute"
	"github.com/ndscm/theseed/seed/cloud/sfe/route/stuffroute"
	"github.com/ndscm/theseed/seed/cloud/sfe/route/workflowroute"
	"github.com/ndscm/theseed/seed/cloud/sfe/signedjwt"
	"github.com/ndscm/theseed/seed/cloud/sqlsession"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/tier/go/servicetier"
	"golang.org/x/sync/errgroup"
)

var flagHttp = seedflag.DefineString("http", "route", "HTTP mode. Available modes: 'route' for normal routing, 'escalate' for escalating all requests, '' for not starting HTTP server.")
var flagHttpPort = seedflag.DefineString("http_port", "9080", "Port for HTTP server")
var flagHttps = seedflag.DefineString("https", "", "HTTPS mode. Available modes: 'route' for normal routing, '' for not starting HTTPS server.")
var flagHttpsPort = seedflag.DefineString("https_port", "9443", "Port for HTTPS server")

var flagSessionProvider = seedflag.DefineString("session_provider", "", "Specify the session provider (e.g., 'sql' for SQL-based sessions)")

var flagSfeOpenidClientId = seedflag.DefineString(
	"openid_client_id", "",
	"Client ID for OpenID Connect",
)
var flagSfeOpenidClientSecret = seedflag.DefineSecret(
	"openid_client_secret",
	"Client Secret for OpenID Connect",
)

var optimizedTransport = &http.Transport{
	Proxy:               nil,
	MaxIdleConns:        200,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     10 * time.Second,
}

type SfeRouteHandler struct {
	parser *servicetier.ServiceTierParser

	golinkRoute http.Handler

	kurisuRoute http.Handler

	stuffRoute http.Handler

	workflowRoute http.Handler
}

func (h *SfeRouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	serviceTier, err := h.parser.Parse(r.Host)
	if err != nil {
		seedlog.Errorf("Unsupported host. err=%v", err)
		http.Error(w, "Bad Gateway", http.StatusBadGateway)
		return
	}
	seedlog.Infof("SFE request: service=%s path=%s", serviceTier, r.URL.Path)

	switch serviceTier.Service {
	case "go":
		h.golinkRoute.ServeHTTP(w, r)
		return
	case "golink":
		h.golinkRoute.ServeHTTP(w, r)
		return
	case "kurisu":
		h.kurisuRoute.ServeHTTP(w, r)
		return
	case "stuff":
		h.stuffRoute.ServeHTTP(w, r)
		return
	case "workflow":
		h.workflowRoute.ServeHTTP(w, r)
		return
	}
	http.Redirect(w, r, "https://www.ndscm.com", http.StatusTemporaryRedirect)
}

func CreateSfeRouteHandler(sfeOpenidClient *openid.OpenidClient) (*SfeRouteHandler, error) {
	sessionInitializer := seedsession.MemorySessionInitializer
	switch flagSessionProvider.Get() {
	case "sql":
		sqlSessionInitializer, err := sqlsession.CreateSqlSessionInitializer()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		sessionInitializer = sqlSessionInitializer
	default:
		seedlog.Warnf("Using in-memory session store, which can not scale for production.")
	}
	golinkRoute, err := golinkroute.CreateGolinkRoute(sfeOpenidClient)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	kurisuRoute, err := kurisuroute.CreateKurisuRoute(sfeOpenidClient)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	stuffRoute, err := stuffroute.CreateStuffRoute(sfeOpenidClient)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	workflowRoute, err := workflowroute.CreateWorkflowRoute(optimizedTransport)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	h := &SfeRouteHandler{
		parser: servicetier.NewStaticParser([]string{"ndscm.com", "ndscm.biz"}),

		golinkRoute:   seedsession.InterceptSessionMiddleware(golinkRoute, sessionInitializer),
		kurisuRoute:   seedsession.InterceptSessionMiddleware(kurisuRoute, sessionInitializer),
		stuffRoute:    seedsession.InterceptSessionMiddleware(stuffRoute, sessionInitializer),
		workflowRoute: workflowRoute,
	}

	return h, nil
}

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithEnvPrefix("SFE_"),
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Service login
	discoveryUrl := openid.OpenidDiscoveryUrlFlag()
	clientId := flagSfeOpenidClientId.Get()
	clientSecret, err := flagSfeOpenidClientSecret.LoadString()
	if err != nil {
		return seederr.Wrap(err)
	}
	sfeOpenidClient, err := signedjwt.WrapOpenidClient(
		discoveryUrl, clientId, clientSecret, nil,
	)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Create routes
	sfeRouteHandler, err := CreateSfeRouteHandler(sfeOpenidClient)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Configure HTTP server
	httpServer := &http.Server{
		Addr: ":" + flagHttpPort.Get(),
	}
	httpMode := flagHttp.Get()
	switch httpMode {
	case "escalate":
		httpServer.Handler = http.HandlerFunc(escalate.ServeHTTP)
	case "route":
		httpServer.Handler = sfeRouteHandler
	default:
		if httpMode != "" {
			seedlog.Warnf("Unknown http mode. httpMode=%s", httpMode)
			httpMode = ""
		}
	}

	// Configure HTTPS server
	sfeCertStore := certstore.NewSfeCertStore(sfeOpenidClient)
	httpsServer := &http.Server{
		Addr: ":" + flagHttpsPort.Get(),
		TLSConfig: &tls.Config{
			GetCertificate: sfeCertStore.GetCertificate,
			NextProtos:     []string{"h2", "http/1.1"},
		},
	}
	httpsMode := flagHttps.Get()
	switch httpsMode {
	case "route":
		httpsServer.Handler = sfeRouteHandler
	default:
		if httpsMode != "" {
			seedlog.Warnf("Unknown https mode. httpsMode=%s", httpsMode)
			httpsMode = ""
		}
	}

	// Start servers
	g, ctx := errgroup.WithContext(context.Background())
	if httpMode != "" {
		g.Go(func() error {
			err := httpServer.ListenAndServe()
			if err != nil {
				return seederr.Wrap(err)
			}
			return nil
		})
	}
	if httpsMode != "" {
		g.Go(func() error {
			err := httpsServer.ListenAndServeTLS("", "")
			if err != nil {
				return seederr.Wrap(err)
			}
			return nil
		})
	}
	g.Go(func() error {
		<-ctx.Done()
		if httpMode != "" {
			err := httpServer.Shutdown(context.Background())
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				seedlog.Errorf("Shutdown http server failed: %v", err)
			}
		}
		if httpsMode != "" {
			err := httpsServer.Shutdown(context.Background())
			if err != nil && !errors.Is(err, http.ErrServerClosed) {
				seedlog.Errorf("Shutdown https server failed: %v", err)
			}
		}
		return ctx.Err()
	})
	err = g.Wait()
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func main() {
	err := run()
	if err != nil {
		seedlog.Errorf("%v", err)
		os.Exit(1)
	}
}
