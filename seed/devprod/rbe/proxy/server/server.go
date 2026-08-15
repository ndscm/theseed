package main

import (
	"bytes"
	"context"
	"encoding/json/v2"
	"errors"
	"io"
	"net"
	"net/http"
	"net/url"
	"os"
	"os/exec"
	"os/signal"
	"path/filepath"
	"strconv"
	"strings"
	"sync"
	"syscall"

	"golang.org/x/oauth2"
	"golang.org/x/sync/errgroup"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/init/go/seedinit"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagRemoteCache = seedflag.DefineString(
	"remote_cache", "",
	"Backend HTTP cache base URL, e.g. https://<bucket>.storage.googleapis.com/ (optional if local_cache is set)",
)
var flagLocalCache = seedflag.DefineString(
	"local_cache", "",
	"Local disk cache directory, checked before the backend and populated from it; usable alone as a local-only cache",
)
var flagLocalPut = seedflag.DefineBool(
	"local_put", false,
	"Accept uploads with no Authorization header by writing them only to the "+
		"local cache, never forwarding them to the backend (requires local_cache). "+
		"When false such uploads are rejected.",
)
var flagListen = seedflag.DefineString(
	"listen", ":8080",
	"Address to listen on: [host]:port or unix:/path/to/socket.path",
)
var flagCredentialHelper = seedflag.DefineString(
	"credential_helper", "",
	"Command run to obtain a backend bearer token; its stdout is interpreted per "+
		"credential_helper_format. Empty leaves the backend unauthenticated except "+
		"for a client-supplied Authorization header.",
)
var flagCredentialHelperFormat = seedflag.DefineString(
	"credential_helper_format", "",
	`How to read the credential_helper output: empty uses stdout as the bearer `+
		`token verbatim; "oauth2" parses stdout as a JSON oauth2.Token and uses its `+
		`access token until it expires, then serializes the token back to the helper `+
		`on stdin to obtain a refreshed one.`,
)

// objectKey maps a request path to its cache object key, rejecting anything that
// is not a well-formed /ac/<hash> or /cas/<hash> key.
func objectKey(urlPath string) (string, bool) {
	key := strings.TrimPrefix(urlPath, "/")
	kind, hash, found := strings.Cut(key, "/")
	if !found {
		return "", false
	}
	if kind != "ac" && kind != "cas" {
		return "", false
	}
	if hash == "" || hash == "." || hash == ".." || strings.Contains(hash, "/") {
		return "", false
	}
	return key, true
}

// localWriter tees cache bytes to a temp file that is atomically renamed into
// place on commit. Local writes are best-effort: a failure is recorded and
// suppressed so it never truncates the client stream or breaks the backend
// request, and commit becomes a no-op.
type localWriter struct {
	path    string
	temp    *os.File
	written int64
	failed  bool
}

func (l *localWriter) Write(p []byte) (int, error) {
	l.written += int64(len(p))
	if l.temp != nil && !l.failed {
		_, err := l.temp.Write(p)
		if err != nil {
			l.failed = true
		}
	}
	return len(p), nil
}

func (l *localWriter) complete(contentLength int64) bool {
	return contentLength < 0 || l.written == contentLength
}

func (l *localWriter) commit() {
	if l.temp == nil {
		return
	}
	name := l.temp.Name()
	err := l.temp.Close()
	if err != nil || l.failed {
		os.Remove(name)
		return
	}
	err = os.Rename(name, l.path)
	if err != nil {
		seedlog.Errorf("Local cache rename failed. path=%s err=%v", l.path, err)
		os.Remove(name)
	}
}

func (l *localWriter) discard() {
	if l.temp == nil {
		return
	}
	name := l.temp.Name()
	l.temp.Close()
	os.Remove(name)
}

// ProxyHandler serves the Bazel HTTP remote cache protocol in the raw /cas, /ac
// layout. With a backend it forwards requests to it (so a GCS backend stays
// interchangeable with a direct --remote_cache=https://<bucket>.storage.googleapis.com/
// configuration); with localCache it also reads through and writes through a
// disk cache in the same layout. Either may be omitted: a backend alone is a
// pure forwarder, a localCache alone is a self-contained local cache.
type ProxyHandler struct {
	remote *url.URL
	client *http.Client

	local string

	// localPut accepts uploads with no Authorization header into the local cache
	// only, without forwarding them to the backend.
	localPut bool
}

// requestRemote issues method against the backend object, attaching the object
// path to the backend base URL. A client-supplied Authorization header is
// forwarded unchanged; otherwise the transport may add a credential-helper
// token.
func (h *ProxyHandler) requestRemote(
	r *http.Request, method string, object string, body io.Reader,
) (*http.Response, error) {
	target := *h.remote
	target.Path = strings.TrimSuffix(h.remote.Path, "/") + "/" + object
	req, err := http.NewRequestWithContext(r.Context(), method, target.String(), body)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	req.ContentLength = r.ContentLength
	if authorization := r.Header.Get("Authorization"); authorization != "" {
		req.Header.Set("Authorization", authorization)
	}
	resp, err := h.client.Do(req)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return resp, nil
}

func (h *ProxyHandler) getLocalObjectPath(object string) string {
	return filepath.Join(h.local, filepath.FromSlash(object))
}

// serveLocalGet serves object from the local cache if it is present, reporting
// whether it did.
func (h *ProxyHandler) serveLocalGet(w http.ResponseWriter, object string) bool {
	if h.local == "" {
		return false
	}
	file, err := os.Open(h.getLocalObjectPath(object))
	if err != nil {
		return false
	}
	defer file.Close()
	info, err := file.Stat()
	if err != nil {
		return false
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
	_, err = io.Copy(w, file)
	if err != nil {
		seedlog.Errorf("Local cache read failed. object=%s err=%v", object, err)
	}
	return true
}

func (h *ProxyHandler) serveGet(w http.ResponseWriter, r *http.Request, object string) {
	if h.serveLocalGet(w, object) {
		return
	}
	if h.remote == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp, err := h.requestRemote(r, http.MethodGet, object, nil)
	if err != nil {
		seedlog.Errorf("Cache upstream request failed. object=%s err=%v", object, err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		w.WriteHeader(resp.StatusCode)
		_, _ = io.Copy(w, resp.Body)
		return
	}

	w.Header().Set("Content-Type", "application/octet-stream")
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
	}

	// Tee the response into the local cache while streaming it to the client,
	// committing only once the full body has been delivered.
	local := h.newLocalWriter(object)
	_, err = io.Copy(w, io.TeeReader(resp.Body, local))
	if err != nil {
		seedlog.Errorf("Cache stream failed. object=%s err=%v", object, err)
		local.discard()
		return
	}
	local.commit()
}

func (h *ProxyHandler) serveHead(w http.ResponseWriter, r *http.Request, object string) {
	if h.local != "" {
		info, err := os.Stat(h.getLocalObjectPath(object))
		if err == nil {
			w.Header().Set("Content-Type", "application/octet-stream")
			w.Header().Set("Content-Length", strconv.FormatInt(info.Size(), 10))
			w.WriteHeader(http.StatusOK)
			return
		}
	}
	if h.remote == nil {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	resp, err := h.requestRemote(r, http.MethodHead, object, nil)
	if err != nil {
		seedlog.Errorf("Cache upstream request failed. object=%s err=%v", object, err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	w.Header().Set("Content-Type", "application/octet-stream")
	if contentLength := resp.Header.Get("Content-Length"); contentLength != "" {
		w.Header().Set("Content-Length", contentLength)
	}
	w.WriteHeader(resp.StatusCode)
}

func (h *ProxyHandler) servePut(w http.ResponseWriter, r *http.Request, object string) {
	// Uploads forwarded to the backend must carry the client's own Authorization.
	// The proxy's credential-helper credential is used only to serve
	// unauthenticated reads (GET/HEAD) from the shared backend; it is never lent
	// to an unauthenticated writer, which could otherwise poison the shared cache
	// anonymously. An unauthenticated upload is instead confined to the local disk
	// cache when localPut is set — never forwarded to the backend —
	// so a local build can still populate the read-through cache; otherwise it is
	// rejected outright. Local-only mode (no backend) has no such credential to
	// protect and always writes locally.
	unauthorized := r.Header.Get("Authorization") == ""
	if h.remote != nil && unauthorized && !h.localPut {
		http.Error(w, "authorization required for uploads", http.StatusForbidden)
		return
	}

	local := h.newLocalWriter(object)

	// Store the body on disk without forwarding when there is no backend, or when
	// an unauthenticated upload is confined to the local cache.
	if h.remote == nil || unauthorized {
		_, err := io.Copy(local, r.Body)
		if err != nil {
			local.discard()
			seedlog.Errorf("Local cache write failed. object=%s err=%v", object, err)
			http.Error(w, "internal error", http.StatusInternalServerError)
			return
		}
		// Commit only a complete object, mirroring the forwarded branch, so a body
		// shorter than its Content-Length never lands in the cache.
		if local.complete(r.ContentLength) {
			local.commit()
		} else {
			local.discard()
		}
		w.WriteHeader(http.StatusNoContent)
		return
	}

	// Tee the uploaded body into the local cache while it is forwarded to the
	// backend, committing only if the backend accepts it.
	resp, err := h.requestRemote(r, http.MethodPut, object, io.TeeReader(r.Body, local))
	if err != nil {
		local.discard()
		seedlog.Errorf("Cache upstream request failed. object=%s err=%v", object, err)
		http.Error(w, "bad gateway", http.StatusBadGateway)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode/100 == 2 && local.complete(r.ContentLength) {
		local.commit()
	} else {
		local.discard()
	}
	w.WriteHeader(resp.StatusCode)
}

func (h *ProxyHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	seedlog.Infof("Cache request. method=%s path=%s", r.Method, r.URL.Path)

	object, ok := objectKey(r.URL.Path)
	if !ok {
		http.Error(w, "not found", http.StatusNotFound)
		return
	}

	switch r.Method {
	case http.MethodGet:
		h.serveGet(w, r, object)
	case http.MethodHead:
		h.serveHead(w, r, object)
	case http.MethodPut:
		h.servePut(w, r, object)
	default:
		w.Header().Set("Allow", "GET, HEAD, PUT")
		http.Error(w, "method not allowed", http.StatusMethodNotAllowed)
	}
}

// newLocalWriter prepares a local cache writer for object. When the local cache
// is disabled or cannot be prepared it returns a no-op writer.
func (h *ProxyHandler) newLocalWriter(object string) *localWriter {
	if h.local == "" {
		return &localWriter{}
	}
	path := h.getLocalObjectPath(object)
	err := os.MkdirAll(filepath.Dir(path), 0o755)
	if err != nil {
		seedlog.Errorf("Local cache mkdir failed. object=%s err=%v", object, err)
		return &localWriter{failed: true}
	}
	temp, err := os.CreateTemp(filepath.Dir(path), ".tmp-*")
	if err != nil {
		seedlog.Errorf("Local cache create failed. object=%s err=%v", object, err)
		return &localWriter{failed: true}
	}
	return &localWriter{path: path, temp: temp}
}

var _ http.Handler = (*ProxyHandler)(nil)

func newProxyHandler(
	remote *url.URL, client *http.Client, localCache string, localPut bool,
) *ProxyHandler {
	return &ProxyHandler{
		remote:   remote,
		client:   client,
		local:    localCache,
		localPut: localPut,
	}
}

// conditionalAuthTransport attaches a bearer token from source, but only when
// the request does not already carry an Authorization header, so a
// client-supplied token is forwarded unchanged.
type conditionalAuthTransport struct {
	source oauth2.TokenSource
	base   http.RoundTripper
}

func (t *conditionalAuthTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	if req.Header.Get("Authorization") == "" {
		token, err := t.source.Token()
		if err != nil {
			return nil, err
		}
		// Clone before mutating: RoundTrip must not modify the caller's request.
		req = req.Clone(req.Context())
		token.SetAuthHeader(req)
	}
	return t.base.RoundTrip(req)
}

var _ http.RoundTripper = (*conditionalAuthTransport)(nil)

// credentialHelperTokenSource obtains a backend bearer token by running an
// external credential helper command. With an empty format the command's stdout
// is the bearer token verbatim and the helper is run for every token, since a
// raw token carries no expiry. With the "oauth2" format the stdout is a
// JSON-encoded oauth2.Token whose access token is reused until it expires; on
// expiry the current token is serialized back to the helper on stdin so it can
// refresh it.
type credentialHelperTokenSource struct {
	command string
	format  string

	mu    sync.Mutex
	token *oauth2.Token
}

// runHelper runs the credential helper command through the shell, writing input
// to its stdin when non-nil, and returns its stdout.
func (s *credentialHelperTokenSource) runHelper(input []byte) ([]byte, error) {
	cmd := exec.Command("sh", "-c", s.command)
	if input != nil {
		cmd.Stdin = bytes.NewReader(input)
	}
	output, err := cmd.Output()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return output, nil
}

func (s *credentialHelperTokenSource) Token() (*oauth2.Token, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Raw format: stdout is the bearer token and carries no expiry, so run the
	// helper for each request rather than caching a token that never lapses.
	if s.format == "" {
		output, err := s.runHelper(nil)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		return &oauth2.Token{AccessToken: strings.TrimSpace(string(output))}, nil
	}

	// oauth2 format: reuse the cached token until it lapses, then refresh it by
	// handing the current token (if any) back to the helper on stdin.
	token := s.token
	if !token.Valid() {
		input := []byte(nil)
		if token != nil {
			marshaled, err := json.Marshal(token)
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			input = marshaled
		}
		output, err := s.runHelper(input)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		refreshed := &oauth2.Token{}
		err = json.Unmarshal(output, refreshed)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		s.token = refreshed
		token = refreshed
	}
	return token, nil
}

var _ oauth2.TokenSource = (*credentialHelperTokenSource)(nil)

// backendTransport returns the transport used for backend requests, optionally
// attaching a credential-helper bearer token when the request has no
// Authorization header. Its base is the default transport, which honors
// HTTPS_PROXY so the backend is reachable through a corporate proxy.
func backendTransport() (http.RoundTripper, error) {
	transport := (http.RoundTripper)(http.DefaultTransport)
	credentialHelper := flagCredentialHelper.Get()
	if credentialHelper != "" {
		format := flagCredentialHelperFormat.Get()
		if format != "" && format != "oauth2" {
			return nil, seederr.WrapErrorf(
				"unknown credential_helper_format: %q (want empty or oauth2)", format)
		}
		transport = &conditionalAuthTransport{
			source: &credentialHelperTokenSource{command: credentialHelper, format: format},
			base:   http.DefaultTransport,
		}
	}
	return transport, nil
}

// listen opens the listener for address, which is either a [host]:port TCP
// address or a unix:/path/to/socket Unix socket. The unix: form matches Bazel's
// --remote_proxy=unix: so the same socket path string can be shared.
func listen(address string) (net.Listener, error) {
	network := "tcp"
	addr := address
	if strings.HasPrefix(address, "unix:") {
		network = "unix"
		addr = strings.TrimPrefix(address, "unix:")
		// Drop any stale socket so the listener can bind it.
		err := os.Remove(addr)
		if err != nil && !errors.Is(err, os.ErrNotExist) {
			return nil, seederr.Wrap(err)
		}
		// Bind the socket owner-only (0700). The socket is unauthenticated, but
		// the proxy behind it holds credential-helper credentials to the shared
		// cache backend, so a world-connectable socket would let any local user
		// read and poison the cache through the proxy. Tightening the umask around
		// the bind
		// avoids the window a post-bind chmod would leave open.
		old := syscall.Umask(0o077)
		defer syscall.Umask(old)
	}
	listener, err := net.Listen(network, addr)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return listener, nil
}

func run() error {
	_, err := seedinit.Initialize(
		seedinit.WithEnvPrefix("RBE_PROXY_"),
		seedinit.WithFallbackEnvPrefix("SEED_"),
	)
	if err != nil {
		return seederr.Wrap(err)
	}

	remoteCache := flagRemoteCache.Get()
	localCache := flagLocalCache.Get()
	if remoteCache == "" && localCache == "" {
		return seederr.WrapErrorf("remote_cache or local_cache is required")
	}

	localPut := flagLocalPut.Get()
	if localPut && localCache == "" {
		return seederr.WrapErrorf("local_put requires local_cache")
	}

	if localCache != "" {
		err := os.MkdirAll(localCache, 0o755)
		if err != nil {
			return seederr.Wrap(err)
		}
	}

	remoteUrl := (*url.URL)(nil)
	client := (*http.Client)(nil)
	if remoteCache != "" {
		remoteUrl, err = url.Parse(remoteCache)
		if err != nil {
			return seederr.Wrap(err)
		}
		if remoteUrl.Scheme != "http" && remoteUrl.Scheme != "https" {
			return seederr.WrapErrorf("remote_cache must be an http(s) URL: %s", remoteCache)
		}
		transport, err := backendTransport()
		if err != nil {
			return seederr.Wrap(err)
		}
		client = &http.Client{Transport: transport}
	}

	handler := newProxyHandler(remoteUrl, client, localCache, localPut)
	httpServer := &http.Server{Handler: handler}

	listener, err := listen(flagListen.Get())
	if err != nil {
		return seederr.Wrap(err)
	}

	// Cancel the root context on SIGINT/SIGTERM so the shutdown goroutine can
	// drain in-flight requests instead of the process being killed outright.
	signalCtx, stop := signal.NotifyContext(context.Background(), os.Interrupt, syscall.SIGTERM)
	defer stop()

	g, ctx := errgroup.WithContext(signalCtx)
	g.Go(func() error {
		seedlog.Infof("RBE cache proxy listening. addr=%s remote_cache=%s local_cache=%s", listener.Addr(), remoteCache, localCache)
		err := httpServer.Serve(listener)
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
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
		return nil
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
