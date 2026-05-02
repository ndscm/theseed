package seedjwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagSkipJwtVerification = seedflag.DefineBool("skip_jwt_verification", false, "Skip JWT verification and trust all JWT (for testing only)")
var flagJwtAudience = seedflag.DefineString("jwt_audience", "", "expected JWT audience (aud claim)")

type JwksStore interface {
	GetByKid(issuer string, kid string) (crypto.PublicKey, error)
}

// See RFC 7515 for JWS structure
// https://datatracker.ietf.org/doc/html/rfc7515
type jwtHeader struct {
	Alg string `json:"alg"`
	Kid string `json:"kid"`
	Typ string `json:"typ"`
}

// See RFC 7519 for JWT claims structure
// https://datatracker.ietf.org/doc/html/rfc7519
type jwtPayload struct {
	Iss string `json:"iss"`

	Aud json.RawMessage `json:"aud"`

	Exp *int64 `json:"exp"`

	Nbf *int64 `json:"nbf"`
}

func (p *jwtPayload) audienceContains(target string) bool {
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

func lookupHashFunction(alg string) (crypto.Hash, error) {
	switch alg {
	case "RS256", "ES256":
		return crypto.SHA256, nil
	case "RS384", "ES384":
		return crypto.SHA384, nil
	case "RS512", "ES512":
		return crypto.SHA512, nil
	default:
		return 0, seederr.WrapErrorf("unsupported JWT algorithm %q", alg)
	}
}

// JwtDecoder verifies OpenID Connect JWTs using statically configured kid-to-certificate mappings.
type JwtDecoder struct {
	audience string

	jwksStore JwksStore
}

func CreateJwtDecoder(jwksStore JwksStore) (*JwtDecoder, error) {
	v := &JwtDecoder{
		audience:  flagJwtAudience.Get(),
		jwksStore: jwksStore,
	}
	return v, nil
}

// resolveSigningKey returns the public key for JWT verification by looking up
// the kid from the pre-loaded trust map.
func (v *JwtDecoder) resolveSigningKey(issuer string, header jwtHeader) (crypto.PublicKey, error) {
	if header.Kid == "" {
		return nil, seederr.WrapErrorf("JWT header has no kid")
	}
	pubKey, err := v.jwksStore.GetByKid(issuer, header.Kid)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	seedlog.Debugf("Jwks resolved kid: %v", header.Kid)
	return pubKey, nil
}

func (v *JwtDecoder) verifyJwtSignature(issuer string, headerB64 string, payloadB64 string, signatureB64 string) error {
	headerBytes, err := base64.RawURLEncoding.DecodeString(headerB64)
	if err != nil {
		return seederr.WrapErrorf("failed to decode JWT header: %v", err)
	}
	header := jwtHeader{}
	err = json.Unmarshal(headerBytes, &header)
	if err != nil {
		return seederr.WrapErrorf("failed to unmarshal JWT header: %v", err)
	}
	seedlog.Debugf("Jwt header: %+v", header)
	if flagSkipJwtVerification.Get() {
		seedlog.Warnf("Skipping JWT verification as it is disabled.")
		return nil
	}
	pubKey, err := v.resolveSigningKey(issuer, header)
	if err != nil {
		return err
	}
	signature, err := base64.RawURLEncoding.DecodeString(signatureB64)
	if err != nil {
		return seederr.WrapErrorf("failed to decode JWT signature: %v", err)
	}
	hashFunc, err := lookupHashFunction(header.Alg)
	if err != nil {
		return err
	}
	h := hashFunc.New()
	h.Write([]byte(headerB64 + "." + payloadB64))
	digest := h.Sum(nil)
	switch key := pubKey.(type) {
	case *rsa.PublicKey:
		err = rsa.VerifyPKCS1v15(key, hashFunc, digest, signature)
		if err != nil {
			return seederr.WrapErrorf("JWT signature verification failed: %v", err)
		}
		seedlog.Debugf("Jwt verified with RSA")
	case *ecdsa.PublicKey:
		keySize := (key.Curve.Params().BitSize + 7) / 8
		if len(signature) != 2*keySize {
			return seederr.WrapErrorf("invalid ECDSA signature length: expected %d, got %d", 2*keySize, len(signature))
		}
		r := new(big.Int).SetBytes(signature[:keySize])
		s := new(big.Int).SetBytes(signature[keySize:])
		if !ecdsa.Verify(key, digest, r, s) {
			return seederr.WrapErrorf("JWT ECDSA signature verification failed")
		}
		seedlog.Debugf("Jwt verified with ECDSA")
	default:
		return seederr.WrapErrorf("unsupported public key type %T", pubKey)
	}
	return nil
}

func (v *JwtDecoder) Decode(accessToken string) ([]byte, error) {
	parts := strings.Split(accessToken, ".")
	if len(parts) != 3 {
		return nil, seederr.WrapErrorf("invalid JWT: expected 3 parts, got %d", len(parts))
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, seederr.WrapErrorf("failed to decode JWT payload: %v", err)
	}
	payload := jwtPayload{}
	err = json.Unmarshal(payloadBytes, &payload)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to unmarshal JWT claims: %v", err)
	}
	err = v.verifyJwtSignature(payload.Iss, parts[0], parts[1], parts[2])
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	now := time.Now().Unix()
	if payload.Exp == nil {
		return nil, seederr.WrapErrorf("JWT missing exp claim")
	}
	if now > *payload.Exp {
		return nil, seederr.WrapErrorf("JWT has expired")
	}
	if payload.Nbf != nil && now < *payload.Nbf {
		return nil, seederr.WrapErrorf("JWT is not yet valid (nbf)")
	}
	if v.audience != "" && !payload.audienceContains(v.audience) {
		return nil, seederr.WrapErrorf("JWT audience does not contain expected %q", v.audience)
	}
	return payloadBytes, nil
}
