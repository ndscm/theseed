package authfe

import (
	"net/http"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type SessionAuthorizationMiddleware struct {
	matcher func(*http.Request) bool

	resolver OpenidProviderResolver

	next http.Handler
}

func (m SessionAuthorizationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()

	if m.matcher != nil && !m.matcher(r) {
		m.next.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	provider, err := m.resolver(r)
	if err != nil {
		seedlog.Errorf("Failed to resolve OpenID provider: %v", err)
		http.Error(w, "Internal Server Error", http.StatusInternalServerError)
		return
	}

	session, err := seedsession.Session(ctx)
	if err != nil {
		seedlog.Warnf("Session error: %v", err)
		m.next.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	tokenSource, err := provider.WrapExternalTokenStorage(ctx, nil, session, nil)
	if err != nil {
		m.next.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	token, err := tokenSource.Token()
	if err != nil || !token.Valid() {
		m.next.ServeHTTP(w, r.WithContext(ctx))
		return
	}

	r.Header.Set("Authorization", "Bearer "+token.AccessToken)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

// Deprecated
func InterceptSessionAuthorizationMiddleware(
	next http.Handler, provider *openid.OpenidProvider,
) http.Handler {
	resolver := func(_ *http.Request) (*openid.OpenidProvider, error) {
		return provider, nil
	}
	return &SessionAuthorizationMiddleware{
		next:     next,
		resolver: resolver,
	}
}

type AuthMiddlewareOptions struct {
	Matcher func(*http.Request) bool
}

func InterceptAuthMiddleware(
	next http.Handler, authHandler *AuthHandler, opts *AuthMiddlewareOptions,
) http.Handler {
	if opts == nil {
		opts = &AuthMiddlewareOptions{}
	}
	return &SessionAuthorizationMiddleware{
		matcher: opts.Matcher,

		resolver: authHandler.resolver,

		next: next,
	}
}
