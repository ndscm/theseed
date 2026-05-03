package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"time"

	"github.com/ndscm/theseed/seed/cloud/sfe/route/golinkroute"
	"github.com/ndscm/theseed/seed/cloud/sqlsession"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/sync/errgroup"
)

var flagHttpPort = seedflag.DefineString("http_port", "80", "")

var flagSessionProvider = seedflag.DefineString("session_provider", "", "Specify the session provider (e.g., 'sql' for SQL-based sessions)")

type SfeHandler struct {
	golinkRoute http.Handler
}

var optimizedTransport = &http.Transport{
	Proxy:               nil,
	MaxIdleConns:        200,
	MaxIdleConnsPerHost: 100,
	IdleConnTimeout:     10 * time.Second,
}

func (h *SfeHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
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
	}
	http.Redirect(w, r, "https://www.ndscm.com", http.StatusTemporaryRedirect)
}

func CreateSfeHandler() (*SfeHandler, error) {
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

	h := &SfeHandler{
		golinkRoute: seedsession.InterceptSessionMiddleware(golinkRoute, sessionInitializer),
	}

	return h, nil
}

func run() error {
	_, err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}

	sfeHandler, err := CreateSfeHandler()
	if err != nil {
		return seederr.Wrap(err)
	}
	httpServer := &http.Server{
		Addr:    ":" + flagHttpPort.Get(),
		Handler: sfeHandler,
	}

	// TODO(nagi): Support HTTPS when sfe server is not behind a proxy that handles TLS termination.

	// Start servers
	g, ctx := errgroup.WithContext(context.Background())
	g.Go(func() error {
		err := httpServer.ListenAndServe()
		if err != nil {
			return seederr.Wrap(err)
		}
		return nil
	})
	g.Go(func() error {
		<-ctx.Done()
		err := httpServer.Shutdown(context.Background())
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			seedlog.Errorf("Shutdown http server failed: %v", err)
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
