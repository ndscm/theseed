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
	"github.com/ndscm/theseed/seed/infra/jwt/go/seedjwt"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func decodeJwt(jwtDecoder *seedjwt.JwtDecoder, accessToken string) (*openid.OpenidUserInfo, error) {
	payload, err := jwtDecoder.Decode(accessToken)
	if err != nil {
		return nil, err
	}
	userInfo, err := openid.DecodeOpenidUserInfo(payload)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return userInfo, nil
}

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

	jwtDecoder *seedjwt.JwtDecoder
}

func (m *openidJwtMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken := strings.TrimPrefix(authorization, "Bearer ")
		userInfo, err := decodeJwt(m.jwtDecoder, accessToken)
		if err != nil {
			seedlog.Errorf("Failed to decode JWT: %v", err)
		}
		if userInfo != nil {
			r = r.WithContext(withOpenidJwtUser(r.Context(), userInfo))
		}
	}
	m.next.ServeHTTP(w, r)
}

type OpenidJwtInterceptor struct {
	jwtDecoder *seedjwt.JwtDecoder
}

func CreateOpenidJwtInterceptor() (*OpenidJwtInterceptor, error) {
	i := &OpenidJwtInterceptor{}
	jwtDecoder, err := seedjwt.CreateJwtDecoder()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	i.jwtDecoder = jwtDecoder
	return i, nil
}

func (i *OpenidJwtInterceptor) Intercept(next http.Handler) http.Handler {
	return &openidJwtMiddleware{next: next, jwtDecoder: i.jwtDecoder}
}
