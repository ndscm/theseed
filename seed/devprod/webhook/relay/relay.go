/*
Relay subscribes to an SSE event stream from a webhook fan-out server and
re-issues each event as an HTTP request to a downstream URL.

Each event's data is the raw HTTP request that was originally POSTed to the
webhook server's broadcast endpoint (produced by httputil.DumpRequest). Relay
parses that back into a request and forwards it, preserving method and headers.

Runs as a single long-lived process. SIGINT/SIGTERM cancels the subscription
and any in-flight relay request, then exits.
*/
package main

import (
	"bufio"
	"bytes"
	"context"
	"io"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/http/go/eventstream"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagEventSource = seedflag.DefineString("event_source", "https://webhook.ndscm.com/github/subscribe", "URL of the event source to connect to")
var flagRelayTo = seedflag.DefineString("relay_to", "https://workflow.ndscm.com/generic-webhook-trigger/invoke", "URL of the event source to relay to")

// relay parses data as a wire-format HTTP request (as written by
// httputil.DumpRequest on the broadcast side) and re-issues it to `to`,
// preserving method and headers. Host and Content-Length are dropped so
// net/http can recompute them for the new destination and body.
//
// The per-call 30s timeout bounds a single relay attempt; the parent ctx
// still governs overall shutdown. Returns an error for transport failures
// or any 4xx/5xx response — the caller decides whether to log and drop or
// retry.
func relay(ctx context.Context, to string, data []byte) error {
	entity, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(data)))
	if err != nil {
		return seederr.Wrap(err)
	}
	body, err := io.ReadAll(entity.Body)
	if err != nil {
		return seederr.Wrap(err)
	}
	entity.Body.Close()

	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	out, err := http.NewRequestWithContext(ctx, entity.Method, to, bytes.NewReader(body))
	if err != nil {
		return seederr.Wrap(err)
	}
	for k, vs := range entity.Header {
		if k == "Host" || k == "Content-Length" {
			continue
		}
		out.Header[k] = vs
	}

	resp, err := http.DefaultClient.Do(out)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer resp.Body.Close()
	io.Copy(io.Discard, resp.Body)
	if resp.StatusCode >= 400 {
		return seederr.WrapErrorf("relay returned %v", resp.Status)
	}
	seedlog.Infof("relayed -> %s %d", to, resp.StatusCode)
	return nil
}

func run() error {
	_, err := seedinit.Initialize()
	if err != nil {
		return seederr.Wrap(err)
	}
	seedlog.Infof("Subscribe to: %v", flagEventSource.Get())
	seedlog.Infof("Relay to: %v", flagRelayTo.Get())

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	err = eventstream.Subscribe(ctx, flagEventSource.Get(), "", func(ev *eventstream.Event) {
		seedlog.Infof("[Event] event:%s id=%s", ev.Event(), ev.Id())
		seedlog.Debugf("data:\n%s", ev.Data())
		err := relay(ctx, flagRelayTo.Get(), ev.Data())
		if err != nil {
			seedlog.Errorf("relay id=%s: %v", ev.Id(), err)
		}
	})
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
