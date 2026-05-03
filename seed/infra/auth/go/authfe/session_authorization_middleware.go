package authfe

import (
	"net/http"

	"github.com/ndscm/theseed/seed/infra/auth/go/loginopenid"
	"github.com/ndscm/theseed/seed/infra/http/go/seedsession"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type SessionAuthorizationMiddleware struct {
	next http.Handler

	provider *loginopenid.UserOpenidProvider
}

func (m SessionAuthorizationMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session, err := seedsession.Session(ctx)
	if err != nil {
		seedlog.Warnf("Session error: %v", err)
		m.next.ServeHTTP(w, r.WithContext(ctx))
		return
	}
	authorization := m.provider.Bearer(ctx, session)
	if authorization != "" {
		r.Header.Set("Authorization", authorization)
	}
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

func InterceptSessionAuthorizationMiddleware(next http.Handler, provider *loginopenid.UserOpenidProvider) http.Handler {
	return &SessionAuthorizationMiddleware{next: next, provider: provider}
}
