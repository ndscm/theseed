package openidverify

import (
	"slices"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/jwt/go/jwsdecoder"
)

var flagTrustOpenidAuthorizedParty = seedflag.DefineStringList(
	"trust_openid_authorized_party", []string{},
	`OAuth 2.0 client IDs trusted as the authorized party (azp claim) of an OpenID token.`,
)

// OpenidDecoder verifies an OpenID Connect token and decodes its claims.
type OpenidDecoder struct {
	// authorizedParties are the client IDs trusted as the token's azp. Empty
	// means every authorized party is accepted.
	authorizedParties []string

	jwsDecoder *jwsdecoder.JwsDecoder
}

// verifyAuthorizedParty checks the azp claim, which names the OAuth 2.0 client
// the token was issued to. When no authorized party is configured, any azp is
// accepted. A token with no azp at all is rejected, as it names no client that
// could be trusted.
func (d *OpenidDecoder) verifyAuthorizedParty(userInfo *openid.OpenidUserInfo) error {
	if len(d.authorizedParties) == 0 {
		return nil
	}
	if userInfo.Azp == "" {
		return seederr.WrapErrorf("OpenID token has no azp claim, expected one of %v", d.authorizedParties)
	}
	if !slices.Contains(d.authorizedParties, userInfo.Azp) {
		return seederr.WrapErrorf("OpenID token azp %q is not a trusted authorized party, expected one of %v", userInfo.Azp, d.authorizedParties)
	}
	return nil
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
	err = d.verifyAuthorizedParty(userInfo)
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
	d := &OpenidDecoder{
		authorizedParties: flagTrustOpenidAuthorizedParty.Get(),

		jwsDecoder: jwsDecoder,
	}
	return d, nil
}
