package jwtcore

import (
	"crypto"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// See RFC 7515 for JWS structure
// https://datatracker.ietf.org/doc/html/rfc7515
type JwsHeader struct {
	// RFC 7515 4.1.1 Algorithm
	Alg string `json:"alg"`

	// RFC 7515 4.1.4 Key ID
	Kid string `json:"kid,omitempty"`

	// RFC 7515 4.1.9 Type
	Typ string `json:"typ"`
}

func (h *JwsHeader) LookupHashFunction() (crypto.Hash, error) {
	switch h.Alg {
	case "RS256", "ES256":
		return crypto.SHA256, nil
	case "RS384", "ES384":
		return crypto.SHA384, nil
	case "RS512", "ES512":
		return crypto.SHA512, nil
	default:
		return 0, seederr.WrapErrorf("unsupported JWT algorithm %q", h.Alg)
	}
}
