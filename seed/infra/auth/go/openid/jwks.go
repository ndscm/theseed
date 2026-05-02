package openid

import (
	"encoding/json"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// OpenidJwk represents a single key in a JWK Set (RFC 7517).
// Only common fields for RSA, EC, and X.509 certificate chain keys are modeled.
type OpenidJwk struct {
	// RFC 7517 4.1
	Kty string `json:"kty"`

	// RFC 7517 4.2
	Use string `json:"use"`

	// RFC 7517 4.4
	Alg string `json:"alg"`

	// RFC 7517 4.5
	Kid string `json:"kid"`

	// RFC 7517 4.7
	X5c []string `json:"x5c"`

	// RFC 7518 6.2.1
	Crv string `json:"crv"`
	X   string `json:"x"`
	Y   string `json:"y"`

	// RFC 7518 6.3.1
	N string `json:"n"`
	E string `json:"e"`
}

type OpenidJwks struct {
	Keys []OpenidJwk `json:"keys"`

	Raw map[string]any `json:"-"`
}

func DecodeOpenidJwks(bytes []byte) (*OpenidJwks, error) {
	jwks := &OpenidJwks{}
	err := json.Unmarshal(bytes, jwks)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = json.Unmarshal(bytes, &jwks.Raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return jwks, nil
}
