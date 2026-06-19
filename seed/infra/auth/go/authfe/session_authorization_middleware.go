package authfe

import (
	"net/http"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type SessionAuthorizationMiddleware struct {
	next http.Handler

	provider *openid.OpenidProvider
}

func (m SessionAuthorizationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, err := seedsession.Session(ctx)
	if err != nil {
		seedlog.Warnf("Session error: %v", err)
		m.next.ServeHTTP(w, r.WithContext(ctx))
		return
	}
	tokenSource, err := m.provider.WrapExternalTokenStorage(ctx, nil, session, nil)
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

func InterceptSessionAuthorizationMiddleware(next http.Handler, provider *openid.OpenidProvider) http.Handler {
	return &SessionAuthorizationMiddleware{next: next, provider: provider}
}
