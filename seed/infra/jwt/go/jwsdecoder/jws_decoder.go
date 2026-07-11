package jwsdecoder

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json"
	"math/big"
	"slices"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/jwt/go/jwtcore"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagSkipJwtVerification = seedflag.DefineBool(
	"skip_jwt_verification", false,
	"Skip JWT verification and trust all JWT (for testing only)",
)
var flagJwtAudience = seedflag.DefineString(
	"jwt_audience", "",
	"the only JWT audience (aud claim) allowed; the JWT must carry exactly this one audience",
)
var flagJwtIncludeAudience = seedflag.DefineStringList(
	"jwt_include_audience", []string{},
	"JWT audiences (aud claim) that must all be present; the JWT may carry additional audiences",
)

// JwsDecoder verifies OpenID Connect JWTs using statically configured kid-to-certificate mappings.
type JwsDecoder struct {
	audience string

	includeAudiences []string

	jwksStore jwtcore.JwksStore
}

// resolveSigningKey returns the public key for JWT verification by looking up
// the kid from the pre-loaded trust map.
func (v *JwsDecoder) resolveSigningKey(issuer string, header jwtcore.JwsHeader) (crypto.PublicKey, error) {
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

func (v *JwsDecoder) verifyJwtSignature(issuer string, headerB64 string, payloadB64 string, signatureB64 string) error {
	headerBytes, err := base64.RawURLEncoding.DecodeString(headerB64)
	if err != nil {
		return seederr.WrapErrorf("failed to decode JWT header: %v", err)
	}
	header := jwtcore.JwsHeader{}
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
	hashFunc, err := header.LookupHashFunction()
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

// verifyAudience checks the aud claim against the configured audience policy.
// When no audience is configured, any audience is accepted.
func (v *JwsDecoder) verifyAudience(aud jwtcore.StringOrStrings) error {
	if v.audience != "" {
		if len(aud) != 1 {
			return seederr.WrapErrorf("JWT must carry exactly 1 audience %q, got %v", v.audience, []string(aud))
		}
		if aud[0] != v.audience {
			return seederr.WrapErrorf("JWT audience is %q, expected %q", aud[0], v.audience)
		}
	}
	for _, includeAudience := range v.includeAudiences {
		if !slices.Contains(aud, includeAudience) {
			return seederr.WrapErrorf("JWT audience does not contain expected %q, got %v", includeAudience, []string(aud))
		}
	}
	return nil
}

func (v *JwsDecoder) Decode(jwsB64 string) ([]byte, error) {
	parts := strings.Split(jwsB64, ".")
	if len(parts) != 3 {
		return nil, seederr.WrapErrorf("invalid JWT: expected 3 parts, got %d", len(parts))
	}
	payloadBytes, err := base64.RawURLEncoding.DecodeString(parts[1])
	if err != nil {
		return nil, seederr.WrapErrorf("failed to decode JWT payload: %v", err)
	}
	payload := jwtcore.JwtPayload{}
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
	err = v.verifyAudience(payload.Aud)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return payloadBytes, nil
}

func CreateJwsDecoder(jwksStore jwtcore.JwksStore) (*JwsDecoder, error) {
	v := &JwsDecoder{
		audience:         flagJwtAudience.Get(),
		includeAudiences: flagJwtIncludeAudience.Get(),

		jwksStore: jwksStore,
	}
	return v, nil
}
