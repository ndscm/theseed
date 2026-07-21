package jwtcore

import (
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// Jwk represents a single key in a JWK Set (RFC 7517).
// Only common fields for RSA, EC, and X.509 certificate chain keys are modeled.
type Jwk struct {
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

type Jwks struct {
	Keys []Jwk `json:"keys"`

	// Inline holds any JWK Set members not matched by a typed field above. The
	// ",inline" option makes json v2 treat it as an inline sink, so a single
	// Unmarshal decodes the known fields and collects the rest here. Each member
	// is kept as raw JSON so its value can be decoded on demand into the type the
	// caller expects, without an eager decode to any that would lose number
	// precision.
	Inline map[string]jsontext.Value `json:",inline"`
}

func DecodeJwks(bytes []byte) (*Jwks, error) {
	jwks := &Jwks{}
	err := json.Unmarshal(bytes, jwks)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return jwks, nil
}
