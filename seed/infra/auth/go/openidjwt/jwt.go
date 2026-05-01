package openidjwt

import (
	"context"
	"encoding/base64"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/ndscm/theseed/seed/infra/context/go/seedctx"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type OpenidUserInfo struct {
	// See: https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
	Sub                 string   `json:"sub"`
	Name                string   `json:"name"`
	GivenName           string   `json:"given_name"`
	FamilyName          string   `json:"family_name"`
	Nickname            string   `json:"nickname"`
	PreferredUsername   string   `json:"preferred_username"`
	Profile             string   `json:"profile"`
	Picture             string   `json:"picture"`
	Website             string   `json:"website"`
	Email               string   `json:"email"`
	EmailVerified       bool     `json:"email_verified"`
	Gender              string   `json:"gender"`
	PhoneNumber         string   `json:"phone_number"`
	PhoneNumberVerified bool     `json:"phone_number_verified"`
	Groups              []string `json:"groups"`

	Raw map[string]interface{} `json:"-"`
}

func decodeJwt(accessToken string) (*OpenidUserInfo, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, seederr.WrapErrorf("invalid JWT: expected 3 parts, got %d", len(parts))
	}
	payload, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, seederr.WrapErrorf("failed to decode JWT payload: %v", err)
	}
	// TODO(nagi): Add signature verification if needed.
	userInfo := &OpenidUserInfo{}
	err = json.Unmarshal(payload, userInfo)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to unmarshal JWT claims: %v", err)
	}
	err = json.Unmarshal(payload, &userInfo.Raw)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to unmarshal JWT claims: %v", err)
	}
	return userInfo, nil
}

func OpenidJwtUser(ctx context.Context) (*OpenidUserInfo, error) {
	if ctx == nil {
		return nil, seederr.WrapErrorf("nil context provided")
	}
	ctxValue := ctx.Value(seedctx.SeedContextKey("openiduser"))
	if ctxValue == nil {
		return nil, nil
	}
	userInfo, ok := ctxValue.(*OpenidUserInfo)
	if !ok {
		return nil, seederr.WrapErrorf("failed to assert user info")
	}
	return userInfo, nil
}

func withOpenidJwtUser(parent context.Context, userInfo *OpenidUserInfo) context.Context {
	return context.WithValue(parent, seedctx.SeedContextKey("openiduser"), userInfo)
}

type openidJwtMiddleware struct {
	next http.Handler
}

func (m *openidJwtMiddleware) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	authorization := r.Header.Get("Authorization")
	if strings.HasPrefix(authorization, "Bearer ") {
		accessToken := strings.TrimPrefix(authorization, "Bearer ")
		userInfo, err := decodeJwt(accessToken)
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
}

func CreateOpenidJwtInterceptor() (*OpenidJwtInterceptor, error) {
	i := &OpenidJwtInterceptor{}
	return i, nil
}

func (i *OpenidJwtInterceptor) Intercept(next http.Handler) http.Handler {
	return &openidJwtMiddleware{next: next}
}
