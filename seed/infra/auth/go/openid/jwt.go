package openid

import (
	"github.com/ndscm/theseed/seed/infra/jwt/go/jwtcore"
)

// OpenidJwtPayload is the claims set of an OpenID Connect ID Token.
// See OpenID Connect Core 1.0 Section 2 for the ID Token definition:
// https://openid.net/specs/openid-connect-core-1_0.html#IDToken
type OpenidJwtPayload struct {
	// The ID Token is a JWT, so it carries the registered JWT claims. Of
	// those, OIDC Core 1.0 Section 2 makes iss, sub, aud, exp and iat
	// REQUIRED.
	jwtcore.JwtPayload

	// OIDC Core 1.0 Section 2 auth_time: time of the End-User
	// authentication, in seconds since the UNIX epoch. REQUIRED when the
	// authentication request used max_age or asked for auth_time as an
	// essential claim, OPTIONAL otherwise.
	AuthTime *int64 `json:"auth_time,omitempty"`

	// OIDC Core 1.0 Section 2 nonce: value passed through unmodified from
	// the authentication request, used to bind the ID Token to a client
	// session and to mitigate replay. Present only if the authentication
	// request carried a nonce, in which case the client MUST check it
	// matches the one it sent.
	Nonce string `json:"nonce,omitempty"`

	// OIDC Core 1.0 Section 2 acr: Authentication Context Class Reference,
	// naming the authentication context class the authentication satisfied.
	// OPTIONAL, and its value is a voluntary claim the client cannot rely on
	// unless it requested it.
	Acr string `json:"acr,omitempty"`

	// OIDC Core 1.0 Section 2 amr: Authentication Methods References,
	// identifiers for the authentication methods used. OPTIONAL, and the
	// identifier values are not defined by the OIDC spec itself.
	Amr []string `json:"amr,omitempty"`

	// OIDC Core 1.0 Section 2 azp: authorized party, the OAuth 2.0 client ID
	// of the party the ID Token was issued to. OPTIONAL, and only needed when
	// the token has a single audience that differs from the authorized party.
	Azp string `json:"azp,omitempty"`
}
