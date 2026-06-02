package main

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ndscm/theseed/seed/cloud/sfe/certstore"
	"github.com/ndscm/theseed/seed/cloud/sfe/escalate"
	"github.com/ndscm/theseed/seed/cloud/sfe/route/golinkroute"
	"github.com/ndscm/theseed/seed/cloud/sfe/route/kurisuroute"
	"github.com/ndscm/theseed/seed/cloud/sfe/route/stuffroute"
	"github.com/ndscm/theseed/seed/cloud/sfe/route/workflowroute"
	"github.com/ndscm/theseed/seed/cloud/sqlsession"
	"github.com/ndscm/theseed/seed/infra/auth/go/clientopenid"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/sync/errgroup"
)

var flagHttp = seedflag.DefineString("http", "route", "HTTP mode. Available modes: 'route' for normal routing, 'escalate' for escalating all requests, '' for not starting HTTP server.")
var flagHttpPort = seedflag.DefineString("http_port", "9080", "Port for HTTP server")
var flagHttps = seedflag.DefineString("https", "", "HTTPS mode. Available modes: 'route' for normal routing, '' for not starting HTTPS server.")
var flagHttpsPort = seedflag.DefineString("https_port", "9443", "Port for HTTPS server")

var flagSessionProvider = seedflag.DefineString("session_provider", "", "Specify the session provider (e.g., 'sql' for SQL-based sessions)")

var flagSfeOpenidClientId = seedflag.DefineString("sfe_openid_client_id", "", "Client ID for OpenID Connect")
var flagSfeOpenidClientSecretFile = seedflag.DefineString("sfe_openid_client_secret_file", "", "Client Secret for OpenID Connect")

var optimizedTransport = &http.Transport{
	Proxy:               nil,
	MaxIdleConns:        200,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     10 * time.Second,
}

type SfeRouteHandler struct {
	golinkRoute http.Handler

	kurisuRoute http.Handler

	stuffRoute http.Handler

	workflowRoute http.Handler
}

func (h *SfeRouteHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	hostname, port, err := net.SplitHostPort(r.Host)
	if err != nil {
		hostname = r.Host
		port = ""
	}
	seedlog.Infof("SFE request: host=%s port=%s path=%s", hostname, port, r.URL.Path)

	switch hostname {
	case "go.ndscm.com":
		h.golinkRoute.ServeHTTP(w, r)
		return
	case "kurisu.ndscm.com":
		h.kurisuRoute.ServeHTTP(w, r)
		return
	case "kurisu.ndscm.biz":
		h.kurisuRoute.ServeHTTP(w, r)
		return
	case "stuff.ndscm.com":
		h.stuffRoute.ServeHTTP(w, r)
		return
	case "workflow.ndscm.biz":
		h.workflowRoute.ServeHTTP(w, r)
		return
	}
	http.Redirect(w, r, "https://www.ndscm.com", http.StatusTemporaryRedirect)
}

func CreateSfeRouteHandler() (*SfeRouteHandler, error) {
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
	golinkRoute, err := golinkroute.CreateGolinkRoute(optimizedTransport)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	kurisuRoute, err := kurisuroute.CreateKurisuRoute(optimizedTransport)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	stuffRoute, err := stuffroute.CreateStuffRoute(optimizedTransport)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	workflowRoute, err := workflowroute.CreateWorkflowRoute(optimizedTransport)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	h := &SfeRouteHandler{
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
	clientId := flagSfeOpenidClientId.Get()
	clientSecret := ""
	clientSecretFile := flagSfeOpenidClientSecretFile.Get()
	if clientSecretFile != "" {
		secretBytes, err := os.ReadFile(clientSecretFile)
		if err != nil {
			return seederr.Wrap(err)
		}
		clientSecret = string(secretBytes)
	}
	serviceOpenid := clientopenid.NewOpenidProvider(
		openid.OpenidDiscoveryUrlFlag(), clientId, clientSecret,
	)

	// Create routes
	sfeRouteHandler, err := CreateSfeRouteHandler()
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
	sfeCertStore := certstore.NewSfeCertStore(serviceOpenid)
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
