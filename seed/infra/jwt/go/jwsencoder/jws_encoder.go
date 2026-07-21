package jwsencoder

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"encoding/json/v2"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/jwt/go/jwtcore"
)

type JwsRsaEncoder struct {
	privateKey *rsa.PrivateKey
}

// Encode builds the JWS compact serialization (RFC 7515) of payload, signed
// with key using the algorithm declared in header. It is the encoding
// counterpart to jwsdecoder.Decode.
func (enc *JwsRsaEncoder) Encode(payload []byte) (string, error) {
	header := jwtcore.JwsHeader{Alg: "RS256", Typ: "JWT"}
	headerBytes, err := json.Marshal(header)
	if err != nil {
		return "", seederr.Wrap(err)
	}

	headerB64 := base64.RawURLEncoding.EncodeToString(headerBytes)
	payloadB64 := base64.RawURLEncoding.EncodeToString(payload)
	input := headerB64 + "." + payloadB64

	hashFunc, err := header.LookupHashFunction()
	if err != nil {
		return "", seederr.Wrap(err)
	}
	h := hashFunc.New()
	h.Write([]byte(input))
	digest := h.Sum(nil)
	signature, err := rsa.SignPKCS1v15(rand.Reader, enc.privateKey, hashFunc, digest)
	if err != nil {
		return "", seederr.Wrap(err)
	}

	jws := input + "." + base64.RawURLEncoding.EncodeToString(signature)
	return jws, nil
}
func (enc *JwsRsaEncoder) EncodeJwt(jwtPayload jwtcore.JwtPayload) (string, error) {
	payloadBytes, err := json.Marshal(jwtPayload)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return enc.Encode(payloadBytes)
}

func CreateJwsRsaEncoder(privateKey *rsa.PrivateKey) *JwsRsaEncoder {
	return &JwsRsaEncoder{
		privateKey: privateKey,
	}
}
