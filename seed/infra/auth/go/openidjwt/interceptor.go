package openidjwt

import (
	"context"
	"net/http"
	"strings"

	_ "crypto/sha256"
	_ "crypto/sha512"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/context/go/seedctx"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func OpenidJwtUser(ctx context.Context) (*openid.OpenidUserInfo, error) {
	if ctx == nil {
		return nil, seederr.WrapErrorf("nil context provided")
	}
	ctxValue := ctx.Value(seedctx.SeedContextKey("openiduser"))
	if ctxValue == nil {
		return nil, nil
	}
	userInfo, ok := ctxValue.(*openid.OpenidUserInfo)
	if !ok {
		return nil, seederr.WrapErrorf("failed to assert user info")
	}
	return userInfo, nil
}

func withOpenidJwtUser(parent context.Context, userInfo *openid.OpenidUserInfo) context.Context {
	return context.WithValue(parent, seedctx.SeedContextKey("openiduser"), userInfo)
}

type openidJwtMiddleware struct {
	next http.Handler

	openidDecoder *OpenidDecoder
}

func (m *openidJwtMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken := strings.TrimPrefix(authorization, "Bearer ")
		userInfo, err := m.openidDecoder.Decode(accessToken)
		if err != nil {
			seedlog.Errorf("Failed to decode JWT: %v", err)
		}
		if userInfo != nil {
			r = r.WithContext(withOpenidJwtUser(r.Context(), userInfo))
		}
	}
	m.next.ServeHTTP(w, r)
}

var _ http.Handler = (*openidJwtMiddleware)(nil)

type OpenidJwtInterceptor struct {
	openidDecoder *OpenidDecoder
}

func (i *OpenidJwtInterceptor) Intercept(next http.Handler) http.Handler {
	return &openidJwtMiddleware{next: next, openidDecoder: i.openidDecoder}
}

func CreateOpenidJwtInterceptor() (*OpenidJwtInterceptor, error) {
	i := &OpenidJwtInterceptor{}
	openidDecoder, err := CreateOpenidDecoder()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	i.openidDecoder = openidDecoder
	return i, nil
}
