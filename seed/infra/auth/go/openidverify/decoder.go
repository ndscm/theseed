package openidverify

import (
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/jwt/go/jwsdecoder"
)

type OpenidDecoder struct {
	jwsDecoder *jwsdecoder.JwsDecoder
}

func (d *OpenidDecoder) Decode(accessToken string) (*openid.OpenidUserInfo, error) {
	payload, err := d.jwsDecoder.Decode(accessToken)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	userInfo, err := openid.DecodeOpenidUserInfo(payload)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return userInfo, nil
}

func CreateOpenidDecoder() (*OpenidDecoder, error) {
	jwksStore, err := CreateOpenidJwksStore()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	jwsDecoder, err := jwsdecoder.CreateJwsDecoder(jwksStore)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return &OpenidDecoder{jwsDecoder: jwsDecoder}, nil
}
