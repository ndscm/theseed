package jwtcore

import (
	"encoding/json"
)

// See RFC 7519 for JWT claims structure
// https://datatracker.ietf.org/doc/html/rfc7519
type JwtPayload struct {
	// RFC 7519 4.1.1 Issuer
	Iss string `json:"iss,omitempty"`

	// RFC 7519 4.1.2 Subject
	Sub string `json:"sub,omitempty"`

	// RFC 7519 4.1.3 Audience
	Aud json.RawMessage `json:"aud,omitempty"`

	// RFC 7519 4.1.4 Expiration Time
	Exp *int64 `json:"exp,omitempty"`

	// RFC 7519 4.1.5 Not Before
	Nbf *int64 `json:"nbf,omitempty"`

	// RFC 7519 4.1.6 Issued At
	Iat *int64 `json:"iat,omitempty"`

	// RFC 7519 4.1.7 JWT ID
	Jti string `json:"jti,omitempty"`
}

func (p *JwtPayload) AudiencesContains(target string) bool {
	single := ""
	if json.Unmarshal(p.Aud, &single) == nil {
		return single == target
	}
	list := []string{}
	if json.Unmarshal(p.Aud, &list) == nil {
		for _, a := range list {
			if a == target {
				return true
			}
		}
	}
	return false
}
