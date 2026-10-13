package authfe

import (
	"context"
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"golang.org/x/oauth2"
)

func sanitizeReturnUrl(raw string) string {
	if raw == "" {
		return "/"
	}
	u, err := url.Parse(raw)
	if err != nil || u.Scheme != "" || u.Host != "" || !strings.HasPrefix(u.Path, "/") {
		return "/"
	}
	return raw
}

func generateState() (string, error) {
	b := make([]byte, 16)
	_, err := rand.Read(b)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return fmt.Sprintf("%x", b), nil
}

func guessOrigin(r *http.Request) string {
	scheme := "http"
	proto := r.Header.Get("X-Forwarded-Proto")
	if proto != "" {
		scheme = proto
	} else if r.TLS != nil {
		scheme = "https"
	}
	return scheme + "://" + r.Host
}

// OpenidProviderResolver returns the OpenID provider to use for a given request. A
// single-tenant front end returns the same provider for every request; a
// multitenant one resolves it per request, e.g. from the request host.
type OpenidProviderResolver func(r *http.Request) (*openid.OpenidProvider, error)

type AuthHandler struct {
	resolver OpenidProviderResolver

	uma bool

	umaAudience string
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/auth/login"):
		h.handleLogin(w, r)
	case strings.HasPrefix(r.URL.Path, "/auth/callback"):
		h.handleCallback(w, r)
	case strings.HasPrefix(r.URL.Path, "/auth/logout"):
		h.handleLogout(w, r)
	case strings.HasPrefix(r.URL.Path, "/auth/clear"):
		h.handleClear(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider, err := h.resolver(r)
	if err != nil {
		seedlog.Errorf("Failed to resolve OpenID provider: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	returnUrl := sanitizeReturnUrl(r.URL.Query().Get("return"))

	session, err := seedsession.Session(ctx)
	if err != nil {
		seedlog.Errorf("Failed to get session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	state, err := generateState()
	if err != nil {
		seedlog.Errorf("Failed to generate state: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	err = session.Update(ctx, map[string]string{
		"oidc_state":      state,
		"oidc_return_url": returnUrl,
	})
	if err != nil {
		seedlog.Errorf("Failed to update session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	origin := guessOrigin(r)
	oauth2Config, err := provider.GetOauth2Config(ctx, origin, nil)
	if err != nil {
		seedlog.Errorf("Failed to get oauth2 config: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	authUrl := oauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, authUrl, http.StatusFound)
}

// grantUmaToken exchanges the freshly obtained access token for a Requesting
// Party Token and persists it to the session, so the stored bearer already
// carries the authorization claim. The audience defaults to the client id, the
// resource server that owns the resources.
func (h *AuthHandler) grantUmaToken(
	ctx context.Context,
	provider *openid.OpenidProvider,
	tokenSource oauth2.TokenSource,
	storage openid.ExternalTokenStorage,
) error {
	token, err := tokenSource.Token()
	if err != nil {
		return seederr.Wrap(err)
	}

	audience := h.umaAudience
	if audience == "" {
		audience = provider.ClientId()
	}

	_, err = provider.UmaGrant(ctx, token.AccessToken, audience, storage)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (h *AuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider, err := h.resolver(r)
	if err != nil {
		seedlog.Errorf("Failed to resolve OpenID provider: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	code := r.URL.Query().Get("code")
	state := r.URL.Query().Get("state")

	if code == "" {
		http.Error(w, "Authorization code not found", http.StatusBadRequest)
		return
	}

	session, err := seedsession.Session(ctx)
	if err != nil {
		seedlog.Errorf("Failed to get session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	savedState, err := session.Get(ctx, "oidc_state")
	if err != nil {
		seedlog.Errorf("Failed to get state from session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	if state == "" || state != savedState {
		seedlog.Warnf("Invalid OIDC state: got %q, expected %q", state, savedState)
		http.Error(w, "Invalid state", http.StatusBadRequest)
		return
	}

	origin := guessOrigin(r)
	tokenSource, err := provider.CodeGrant(ctx, origin, code, nil, session)
	if err != nil {
		seedlog.Errorf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
	}

	if h.uma {
		err = h.grantUmaToken(ctx, provider, tokenSource, session)
		if err != nil {
			// The code grant already stored the plain access token, so unprotected
			// routes still work; the protected ones are refused later all the same.
			seedlog.Warnf("Failed to grant UMA token, keeping plain access token: %v", err)
		}
	}

	returnUrlRaw, err := session.Get(ctx, "oidc_return_url")
	if err != nil {
		returnUrlRaw = ""
	}
	returnUrl := sanitizeReturnUrl(returnUrlRaw)

	// Set auth state to empty rather than deleting keys; key removal in a
	// concurrent-safe JSON column is error-prone. Expired sessions are purged
	// entirely by the external cleanup process.
	err = session.Update(ctx, map[string]string{
		"oidc_state":      "",
		"oidc_return_url": "",
	})
	if err != nil {
		seedlog.Warnf("Failed to clear auth state from session: %v", err)
	}

	http.Redirect(w, r, returnUrl, http.StatusFound)
}

// handleLogout clears the local grant and redirects to the provider's
// end-session endpoint to end its session too.
func (h *AuthHandler) handleLogout(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider, err := h.resolver(r)
	if err != nil {
		seedlog.Errorf("Failed to resolve OpenID provider: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	returnUrl := sanitizeReturnUrl(r.URL.Query().Get("return"))
	redirectUrl := guessOrigin(r) + returnUrl

	session, err := seedsession.Session(ctx)
	if err != nil {
		seedlog.Errorf("Failed to get session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	idToken, err := provider.ClearGrant(ctx, session)
	if err != nil {
		seedlog.Errorf("Failed to clear grant: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	endSessionUrl, err := provider.EndSessionUrl(ctx, &openid.EndSessionOptions{
		IdTokenHint: idToken,

		PostLogoutRedirectUri: redirectUrl,
	})
	if err != nil {
		seedlog.Warnf("Failed to build end-session URL: %v", err)
		endSessionUrl = ""
	}

	if endSessionUrl == "" {
		endSessionUrl = returnUrl
	}
	http.Redirect(w, r, endSessionUrl, http.StatusFound)
}

// handleClear clears only the local grant, without contacting the provider,
// for when the provider-side identity changed but the front end holds the old
// session.
func (h *AuthHandler) handleClear(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	provider, err := h.resolver(r)
	if err != nil {
		seedlog.Errorf("Failed to resolve OpenID provider: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	returnUrl := sanitizeReturnUrl(r.URL.Query().Get("return"))

	session, err := seedsession.Session(ctx)
	if err != nil {
		seedlog.Errorf("Failed to get session: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	_, err = provider.ClearGrant(ctx, session)
	if err != nil {
		seedlog.Errorf("Failed to clear stored tokens: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	http.Redirect(w, r, returnUrl, http.StatusFound)
}

type AuthHandlerOptions struct {
	Uma bool

	UmaAudience string
}

type AuthHandlerOption func(*AuthHandlerOptions)

func WithUma(audience string) AuthHandlerOption {
	return func(opts *AuthHandlerOptions) {
		opts.Uma = true
		opts.UmaAudience = audience
	}
}

// NewAuthHandler builds a single-tenant auth handler bound to a fixed provider.
// For a multitenant front end that resolves the provider per request, use
// NewMultiTenantAuthHandler instead.
func NewAuthHandler(provider *openid.OpenidProvider, opts ...AuthHandlerOption) *AuthHandler {
	options := &AuthHandlerOptions{}
	for _, opt := range opts {
		opt(options)
	}
	resolver := func(_ *http.Request) (*openid.OpenidProvider, error) {
		return provider, nil
	}
	return &AuthHandler{
		resolver: resolver,

		uma: options.Uma,

		umaAudience: options.UmaAudience,
	}
}

// NewMultiTenantAuthHandler builds an auth handler that resolves its OpenID
// provider per request via resolver, for front ends that serve multiple tenants
// from a single process.
func NewMultiTenantAuthHandler(resolver OpenidProviderResolver, opts *AuthHandlerOptions) *AuthHandler {
	if opts == nil {
		opts = &AuthHandlerOptions{}
	}
	return &AuthHandler{
		resolver: resolver,

		uma: opts.Uma,

		umaAudience: opts.UmaAudience,
	}
}
