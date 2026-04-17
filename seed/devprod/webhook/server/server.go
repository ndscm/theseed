/*
Server is a webhook fan-out service intended for single-instance deployment.

All state is held in process memory and is not shared, replicated, or
persisted across restarts. Running multiple instances behind a load
balancer is not supported.
*/
package main

import (
	"context"
	"errors"
	"net"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/ndscm/theseed/seed/devprod/webhook/go/webhook"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagPort = seedflag.DefineString("port", "4665", "Port to run the server on") // Default port assignment word: HOOK (4665)
var flagChannels = seedflag.DefineString("channels", "/broadcast:/subscribe", "Comma separated list, each item in the form /<broadcast>:/<subscribe>")

// interceptBodyLimitMiddleware caps request body size at max bytes. Required by
// webhook.NewWebhookHandler, which buffers entire POST bodies in memory — without
// this, an unbounded upload would OOM the process.
func interceptBodyLimitMiddleware(h http.Handler, max int64) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		r.Body = http.MaxBytesReader(w, r.Body, max)
		h.ServeHTTP(w, r)
	})
}

func run() error {
	err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}

	mux := http.NewServeMux()
	channels := strings.Split(flagChannels.Get(), ",")
	for _, channel := range channels {
		// SplitN with n=3 so inputs with extra colons (e.g. "/a:/b:/c") land in
		// the len != 2 branch below and are rejected rather than silently
		// treated as "/a" + "/b:/c".
		endpoints := strings.SplitN(channel, ":", 3)
		if len(endpoints) != 2 || !strings.HasPrefix(endpoints[0], "/") || !strings.HasPrefix(endpoints[1], "/") {
			return seederr.WrapErrorf("invalid channel %q, expected /<broadcast>:/<subscribe>", channel)
		}
		broadcastPath := endpoints[0]
		subscribePath := endpoints[1]
		handler := webhook.NewWebhookHandler(broadcastPath, subscribePath)
		// Panics if broadcastPath or subscribePath collides with a prior channel.
		mux.Handle(broadcastPath, handler)
		mux.Handle(subscribePath, handler)
		seedlog.Infof("Registered webhook channel: broadcast=%s subscribe=%s", broadcastPath, subscribePath)
	}
	handler := interceptBodyLimitMiddleware(mux, 8<<20)

	// ctx is cancelled on SIGINT/SIGTERM. Plumbed into the server via
	// BaseContext below so in-flight requests (especially long-lived SSE
	// subscribers) see cancellation via r.Context() and exit cleanly —
	// otherwise srv.Shutdown would block on them until the timeout.
	ctx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	srv := &http.Server{
		Addr:              ":" + flagPort.Get(),
		Handler:           handler,
		ReadHeaderTimeout: 10 * time.Second,
		BaseContext:       func(net.Listener) context.Context { return ctx },
	}

	// Shutdown runs in a goroutine so it can overlap with ListenAndServe's
	// return. Buffer of 1 lets the goroutine send and exit even if this
	// function returns early (e.g. bind failure) without reading.
	shutdownErr := make(chan error, 1)
	go func() {
		<-ctx.Done()
		seedlog.Infof("Shutting down webhook server")
		shutdownCtx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		shutdownErr <- srv.Shutdown(shutdownCtx)
	}()

	seedlog.Infof("Starting webhook server on %s", srv.Addr)
	err = srv.ListenAndServe()

	// ErrServerClosed is only returned after Shutdown/Close. We never call
	// Close, so it implies the shutdown goroutine ran and will write to
	// shutdownErr. Gating the channel read on `closed` prevents a deadlock
	// on the (spec-impossible) path where ListenAndServe returns nil.
	closed := errors.Is(err, http.ErrServerClosed)
	if err != nil && !closed {
		return seederr.Wrap(err)
	}
	if closed {
		err = <-shutdownErr
		if err != nil {
			return seederr.Wrap(err)
		}
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
