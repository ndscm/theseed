package webhook

import (
	"bytes"
	"net/http"
	"net/http/httputil"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

const maxSubscribers = 32

type subscriber struct {
	ch chan []byte
}

type webhookHandler struct {
	broadcastPath string
	subscribePath string

	mu          sync.Mutex
	subscribers map[*subscriber]struct{}
	nextId      atomic.Uint64
}

// NewWebhookHandler returns a webhook fan-out handler intended for
// single-instance deployment: subscriber set, event ids, and any future replay
// buffer live in process memory and are not shared or persisted across restarts.
//
// broadcast() buffers the entire request body in memory via httputil.DumpRequest,
// so an unbounded POST would OOM the process. Callers must wrap this handler
// with body-limiting middleware (http.MaxBytesReader) and serve it from an
// http.Server configured with ReadTimeout and MaxHeaderBytes.
func NewWebhookHandler(broadcastPath string, subscribePath string) *webhookHandler {
	// TODO(nagi): add external message queue support for scalability and durability.
	return &webhookHandler{
		broadcastPath: broadcastPath,
		subscribePath: subscribePath,
		subscribers:   make(map[*subscriber]struct{}),
	}
}

func (h *webhookHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case r.URL.Path == h.broadcastPath:
		// Accept every http method for broadcast.
		h.broadcast(w, r)
		return
	case r.URL.Path == h.subscribePath:
		if r.Method == http.MethodGet {
			h.subscribe(w, r)
			return
		}
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
		return
	}
	http.Error(w, "not found", http.StatusNotFound)
}

func (h *webhookHandler) broadcast(w http.ResponseWriter, r *http.Request) {
	// TODO(nagi): verify webhook signature.
	seedlog.Infof("Received %v event from: %v", r.Method, r.RemoteAddr)
	r.Header.Del("Authorization")
	r.Header.Del("Cookie")
	raw, err := httputil.DumpRequest(r, true)
	if err != nil {
		http.Error(w, err.Error(), http.StatusBadRequest)
		return
	}

	id := h.nextId.Add(1)

	frame := bytes.Buffer{}
	frame.WriteString("id: ")
	frame.WriteString(strconv.FormatUint(id, 10))
	frame.WriteByte('\n')
	for line := range bytes.SplitSeq(raw, []byte{'\n'}) {
		line = bytes.TrimSuffix(line, []byte{'\r'})
		frame.WriteString("data: ")
		frame.Write(line)
		frame.WriteByte('\n')
	}
	frame.WriteByte('\n')
	msg := frame.Bytes()

	h.mu.Lock()
	targets := make([]*subscriber, 0, len(h.subscribers))
	for s := range h.subscribers {
		targets = append(targets, s)
	}
	h.mu.Unlock()

	dropped := 0
	for _, s := range targets {
		select {
		case s.ch <- msg:
		default:
			dropped++
		}
	}
	if dropped > 0 {
		seedlog.Warnf("dropped event id=%d for %d slow subscriber(s)", id, dropped)
	}
	w.WriteHeader(http.StatusNoContent)
}

func (h *webhookHandler) subscribe(w http.ResponseWriter, r *http.Request) {
	// TODO(nagi): honor Last-Event-ID by replaying from a bounded ring buffer.
	flusher, ok := w.(http.Flusher)
	if !ok {
		http.Error(w, "streaming not supported", http.StatusInternalServerError)
		return
	}

	s := &subscriber{ch: make(chan []byte, 32)}
	h.mu.Lock()
	if len(h.subscribers) >= maxSubscribers {
		h.mu.Unlock()
		seedlog.Warnf("subscriber rejected from %v: cap=%d reached", r.RemoteAddr, maxSubscribers)
		http.Error(w, "too many subscribers", http.StatusServiceUnavailable)
		return
	}
	h.subscribers[s] = struct{}{}
	count := len(h.subscribers)
	h.mu.Unlock()

	w.Header().Set("Content-Type", "text/event-stream")
	w.Header().Set("Cache-Control", "no-cache")
	w.Header().Set("Connection", "keep-alive")
	w.WriteHeader(http.StatusOK)
	flusher.Flush()

	seedlog.Infof("subscriber connected from %v (total=%d)", r.RemoteAddr, count)

	defer func() {
		h.mu.Lock()
		delete(h.subscribers, s)
		remaining := len(h.subscribers)
		h.mu.Unlock()
		seedlog.Infof("subscriber disconnected from %v (total=%d)", r.RemoteAddr, remaining)
	}()

	keepalive := time.NewTicker(30 * time.Second)
	defer keepalive.Stop()

	ctx := r.Context()
	for {
		select {
		case <-ctx.Done():
			return
		case msg := <-s.ch:
			_, err := w.Write(msg)
			if err != nil {
				return
			}
			flusher.Flush()
		case <-keepalive.C:
			_, err := w.Write([]byte(": ping\n\n"))
			if err != nil {
				return
			}
			flusher.Flush()
		}
	}
}
