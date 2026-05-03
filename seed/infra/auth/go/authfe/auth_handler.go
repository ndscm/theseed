package authfe

import (
	"crypto/rand"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/ndscm/theseed/seed/infra/auth/go/loginopenid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
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

type AuthHandler struct {
	provider *loginopenid.UserOpenidProvider
}

func (h *AuthHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	switch {
	case strings.HasPrefix(r.URL.Path, "/auth/login"):
		h.handleLogin(w, r)
	case strings.HasPrefix(r.URL.Path, "/auth/callback"):
		h.handleCallback(w, r)
	default:
		http.NotFound(w, r)
	}
}

func (h *AuthHandler) handleLogin(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	origin := "https://" + r.Host
	oauth2Config, err := h.provider.GetOauth2Config(ctx, origin)
	if err != nil {
		seedlog.Errorf("Failed to get oauth2 config: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	authUrl := oauth2Config.AuthCodeURL(state)
	http.Redirect(w, r, authUrl, http.StatusFound)
}

func (h *AuthHandler) handleCallback(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

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

	origin := "https://" + r.Host
	err = h.provider.Exchange(ctx, session, origin, code)
	if err != nil {
		seedlog.Errorf("Failed to exchange code: %v", err)
		http.Error(w, "Failed to exchange authorization code", http.StatusInternalServerError)
		return
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

func NewAuthHandler(provider *loginopenid.UserOpenidProvider) *AuthHandler {
	return &AuthHandler{provider: provider}
}
