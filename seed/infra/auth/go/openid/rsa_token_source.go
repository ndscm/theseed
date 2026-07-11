package openid

import (
	"crypto/rand"
	"crypto/rsa"
	"encoding/base64"
	"time"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/jwt/go/jwsencoder"
	"github.com/ndscm/theseed/seed/infra/jwt/go/jwtcore"
	"golang.org/x/oauth2"
)

type RsaTokenSource struct {
	clientId string

	audiences []string

	jwsEncoder *jwsencoder.JwsRsaEncoder
}

func (s *RsaTokenSource) Token() (*oauth2.Token, error) {
	now := time.Now()
	exp := now.Add(time.Minute)

	jtiBytes := make([]byte, 16)
	_, err := rand.Read(jtiBytes)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	iat := now.Unix()
	expUnix := exp.Unix()
	payload := jwtcore.JwtPayload{
		Iss: s.clientId,
		Sub: s.clientId,
		Aud: s.audiences,
		Jti: base64.RawURLEncoding.EncodeToString(jtiBytes),
		Iat: &iat,
		Exp: &expUnix,
	}

	accessToken, err := s.jwsEncoder.EncodeJwt(payload)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	token := &oauth2.Token{
		AccessToken: accessToken,
		Expiry:      exp,
	}
	return token, nil
}

var _ oauth2.TokenSource = (*RsaTokenSource)(nil)

func NewRsaTokenSource(clientId string, audiences []string, privateKey *rsa.PrivateKey) *RsaTokenSource {
	encoder := jwsencoder.CreateJwsRsaEncoder(privateKey)
	return &RsaTokenSource{
		clientId:   clientId,
		audiences:  audiences,
		jwsEncoder: encoder,
	}
}
