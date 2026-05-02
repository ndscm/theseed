package seedjwt

import (
	"crypto"
	"crypto/ecdsa"
	"crypto/rsa"
	"crypto/x509"
	"encoding/base64"
	"encoding/json"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

var flagStaticJwks = seedflag.DefineString("static_jwks", "/etc/seed/jwks.json", `JWT trust config file.
Set to empty string to dangerously trust all kid-bearing JWTs without verification (not recommended).`)
var flagJwtIssuer = seedflag.DefineString("jwt_issuer", "", "expected JWT issuer (iss claim)")
var flagJwtAudience = seedflag.DefineString("jwt_audience", "", "expected JWT audience (aud claim)")

// staticJwksConfig maps each JWT kid to its certificate file path.
type staticJwksConfig map[string]string

type JwksStore interface {
	GetByKid(issuer string, kid string) (crypto.PublicKey, error)
}

type staticJwksStore struct {
	certificates map[string]*x509.Certificate
}

func (s *staticJwksStore) GetByKid(issuer string, kid string) (crypto.PublicKey, error) {
	whitelistIssuer := flagJwtIssuer.Get()
	if whitelistIssuer != "" && issuer != whitelistIssuer {
		return nil, seederr.WrapErrorf("JWT issuer %q does not match expected %q", issuer, whitelistIssuer)
	}
	cert, ok := s.certificates[kid]
	if !ok {
		return nil, seederr.WrapErrorf("certificate not found for kid %v", kid)
	}
	now := time.Now()
	if now.Before(cert.NotBefore) {
		return nil, seederr.WrapErrorf("certificate %v is not yet valid (NotBefore: %v)", kid, cert.NotBefore)
	}
	if now.After(cert.NotAfter) {
		return nil, seederr.WrapErrorf("certificate %v has expired (NotAfter: %v)", kid, cert.NotAfter)
	}
	return cert.PublicKey, nil
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

func loadCertFile(certPath string) (*x509.Certificate, error) {
	certPEM, err := os.ReadFile(certPath)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to read trust certificate %v: %v", certPath, err)
	}
	block, _ := pem.Decode(certPEM)
	if block == nil {
		return nil, seederr.WrapErrorf("failed to decode PEM block from %v", certPath)
	}
	cert, err := x509.ParseCertificate(block.Bytes)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to parse certificate from %v: %v", certPath, err)
	}
	now := time.Now()
	if now.Before(cert.NotBefore) {
		seedlog.Warnf("Certificate %v is not yet valid (NotBefore: %v)", certPath, cert.NotBefore)
	}
	if now.After(cert.NotAfter) {
		seedlog.Warnf("Certificate %v has expired (NotAfter: %v)", certPath, cert.NotAfter)
	}
	return cert, nil
}

// JwtDecoder verifies OpenID Connect JWTs using statically configured kid-to-certificate mappings.
type JwtDecoder struct {
	dangerousTrustAllKid bool

	audience string

	jwksStore JwksStore
}

func CreateJwtDecoder() (*JwtDecoder, error) {
	v := &JwtDecoder{
		audience: flagJwtAudience.Get(),
	}
	// TODO(nagi): Add support for dynamic JWKS fetching and caching.
	configPath := flagStaticJwks.Get()
	if configPath == "" {
		v.dangerousTrustAllKid = true
		return v, nil
	}
	configBytes, err := os.ReadFile(configPath)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to read jwks config %v: %v", configPath, err)
	}
	jwksConfig := staticJwksConfig{}
	err = json.Unmarshal(configBytes, &jwksConfig)
	if err != nil {
		return nil, seederr.WrapErrorf("failed to parse jwks config %v: %v", configPath, err)
	}
	staticStore := &staticJwksStore{certificates: map[string]*x509.Certificate{}}
	for kid, certPath := range jwksConfig {
		if !strings.HasPrefix(certPath, "/") {
			configAbsPath, err := filepath.Abs(configPath)
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			configDir := filepath.Dir(configAbsPath)
			certPath = filepath.Join(configDir, certPath)
		}
		cert, err := loadCertFile(certPath)
		if err != nil {
			return nil, seederr.WrapErrorf("failed to load certificate for kid %q: %v", kid, err)
		}
		staticStore.certificates[kid] = cert
	}
	v.jwksStore = staticStore
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
	if v.dangerousTrustAllKid && header.Kid != "" {
		seedlog.Warnf("Skipping JWT verification as it is disabled. Received kid: %v", header.Kid)
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
