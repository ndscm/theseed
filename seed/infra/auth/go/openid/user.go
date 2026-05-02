package openid

import (
	"encoding/json"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type OpenidUserInfo struct {
	// See: https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
	Sub                 string   `json:"sub"`
	Name                string   `json:"name"`
	GivenName           string   `json:"given_name"`
	FamilyName          string   `json:"family_name"`
	Nickname            string   `json:"nickname"`
	PreferredUsername   string   `json:"preferred_username"`
	Profile             string   `json:"profile"`
	Picture             string   `json:"picture"`
	Website             string   `json:"website"`
	Email               string   `json:"email"`
	EmailVerified       bool     `json:"email_verified"`
	Gender              string   `json:"gender"`
	PhoneNumber         string   `json:"phone_number"`
	PhoneNumberVerified bool     `json:"phone_number_verified"`
	Groups              []string `json:"groups"`

	Raw map[string]any `json:"-"`
}

func DecodeOpenidUserInfo(bytes []byte) (*OpenidUserInfo, error) {
	userInfo := &OpenidUserInfo{}
	err := json.Unmarshal(bytes, userInfo)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = json.Unmarshal(bytes, &userInfo.Raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return userInfo, nil
}
