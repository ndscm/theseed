package seedsession

import (
	"context"
	"net/http"
	"time"

	"github.com/ndscm/theseed/seed/infra/context/go/seedctx"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

const SessionIdCookieName = "SID"

func WrapCookieString(sessionId string, expires time.Time) string {
	c := http.Cookie{
		Name:     SessionIdCookieName,
		Value:    sessionId,
		Expires:  expires,
		HttpOnly: true,
		Secure:   true,
		SameSite: http.SameSiteNoneMode,
	}
	return c.String()
}

type SessionAdapter interface {
	SessionId() string

	Init(ctx context.Context, sessionId string, responseHeaders http.Header) error
	Refresh(ctx context.Context, responseHeaders http.Header) error
	Reload(ctx context.Context) error
	Get(ctx context.Context, key string) (string, error)
	Update(ctx context.Context, key string, value string) error
}

func Session(ctx context.Context) (SessionAdapter, error) {
	if ctx == nil {
		return nil, seederr.WrapErrorf("nil context provided")
	}
	session, ok := ctx.Value(seedctx.MptContextKey("session")).(SessionAdapter)
	if !ok {
		return nil, seederr.WrapErrorf("session not found in context")
	}
	return session, nil
}

func withSession(parent context.Context, session SessionAdapter) context.Context {
	return context.WithValue(parent, seedctx.MptContextKey("session"), session)
}

type sessionMiddleware struct {
	next               http.Handler
	sessionInitializer func() SessionAdapter
}

// ServeHTTP implements the http.Handler interface.
func (m sessionMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	ctx := r.Context()
	session := m.sessionInitializer()
	cookieSessionId, err := r.Cookie(SessionIdCookieName)
	if err != nil && err != http.ErrNoCookie {
		seedlog.Errorf("Failed to read cookie: %v", err)
		http.Error(w, "Failed to read cookie", http.StatusBadRequest)
		return
	}
	sessionId := ""
	if err == nil {
		if cookieSessionId != nil {
			sessionId = cookieSessionId.Value
		}
	}
	err = session.Init(ctx, sessionId, w.Header())
	if err != nil {
		seedlog.Errorf("Failed to initialize session: %v", err)
		http.Error(w, "Failed to initialize session", http.StatusInternalServerError)
		return
	}
	ctx = withSession(ctx, session)
	m.next.ServeHTTP(w, r.WithContext(ctx))
}

func InterceptSessionMiddleware(next http.Handler, sessionInitializer func() SessionAdapter) http.Handler {
	return &sessionMiddleware{next: next, sessionInitializer: sessionInitializer}
}
