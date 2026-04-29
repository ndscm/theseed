package seedbearer

import (
	"context"
	"net/http"
	"strings"

	"github.com/ndscm/theseed/seed/infra/context/go/seedctx"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

func Bearer(ctx context.Context) (string, error) {
	if ctx == nil {
		return "", seederr.WrapErrorf("nil context provided")
	}
	token, ok := ctx.Value(seedctx.SeedContextKey("bearer")).(string)
	if !ok {
		return "", nil
	}
	return token, nil
}

func WithBearer(parent context.Context, token string) context.Context {
	return context.WithValue(parent, seedctx.SeedContextKey("bearer"), token)
}

type bearerMiddleware struct {
	next http.Handler
}

func (g *bearerMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		token := strings.TrimPrefix(authorization, "Bearer ")
		if token != "" {
			r = r.WithContext(WithBearer(r.Context(), token))
		}
	}
	g.next.ServeHTTP(w, r)
}

func InterceptBearerMiddleware(next http.Handler) http.Handler {
	return &bearerMiddleware{next: next}
}

type bearerTransport struct {
	next http.RoundTripper
}

func (t *bearerTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	token, _ := Bearer(req.Context())
	if token != "" {
		req = req.Clone(req.Context())
		req.Header.Set("Authorization", "Bearer "+token)
	}
	return t.next.RoundTrip(req)
}

func InterceptBearerTransport(next *http.Client) *http.Client {
	if next == nil {
		next = http.DefaultClient
	}
	if next.Transport == nil {
		next.Transport = http.DefaultTransport
	}
	next.Transport = &bearerTransport{next: next.Transport}
	return next
}
