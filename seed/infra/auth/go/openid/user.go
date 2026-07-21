package openid

import (
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

// OpenidUserInfo describes an End-User with the OpenID Connect standard claims.
// See OpenID Connect Core 1.0 Section 5.1 for the standard claim definitions:
// https://openid.net/specs/openid-connect-core-1_0.html#StandardClaims
//
// The claims may arrive either in an ID Token or in a UserInfo Endpoint
// response. An OpenID Provider is only obliged to return the claims the client
// asked for and consent was given for, so every claim other than sub is best
// treated as absent until proven otherwise.
type OpenidUserInfo struct {
	// The standard claims can be carried by an ID Token, which also carries the
	// ID Token claims of OIDC Core 1.0 Section 2. The sub claim is one of them,
	// and is promoted from here rather than redeclared: it is stable and never
	// reassigned, but is unique only within the scope of the Issuer, so an
	// End-User is identified by the Iss and Sub pair rather than by Sub alone.
	OpenidJwtPayload

	// name: full name in displayable form, including all name parts and
	// possibly titles and suffixes, ordered per the End-User's locale.
	Name string `json:"name,omitempty"`

	// given_name: given name(s) or first name(s). Some cultures allow several,
	// in which case they are all present, separated by spaces.
	GivenName string `json:"given_name,omitempty"`

	// family_name: surname(s) or last name(s). Some cultures allow several or
	// none; several are all present, separated by spaces.
	FamilyName string `json:"family_name,omitempty"`

	// nickname: casual name, which may or may not match GivenName. For example
	// a nickname of "Mike" alongside a given name of "Michael".
	Nickname string `json:"nickname,omitempty"`

	// preferred_username: shorthand name the End-User wishes to be referred to
	// by, such as "janedoe" or "j.doe". It may be any JSON string, including
	// one with "@", "/" or whitespace, and the spec explicitly warns clients
	// MUST NOT rely on it being unique.
	PreferredUsername string `json:"preferred_username,omitempty"`

	// profile: URL of the End-User's profile page.
	Profile string `json:"profile,omitempty"`

	// picture: URL of the End-User's profile picture. It points at an image
	// file, not at a page containing an image.
	Picture string `json:"picture,omitempty"`

	// website: URL of the End-User's web page or blog.
	Website string `json:"website,omitempty"`

	// email: preferred e-mail address, in RFC 5322 addr-spec syntax. The spec
	// explicitly warns clients MUST NOT rely on it being unique, so it is not
	// safe to use as an account key on its own.
	Email string `json:"email,omitempty"`

	// email_verified: whether the Provider took affirmative steps to establish
	// that the End-User controlled Email at the time it was verified. What
	// counts as verification is up to the Provider, so how much this is worth
	// depends on which Provider issued it. A nil value means the Provider did
	// not send the claim, which is not the same as it sending false.
	EmailVerified *bool `json:"email_verified,omitempty"`

	// gender: the values the spec defines are "female" and "male", but a
	// Provider MAY send any other value when neither applies.
	Gender string `json:"gender,omitempty"`

	// phone_number: preferred telephone number, RECOMMENDED to be in E.164
	// format, for example "+1 (425) 555-1212", with any extension in the RFC
	// 3966 syntax, for example "+1 (604) 555-1234;ext=5678".
	PhoneNumber string `json:"phone_number,omitempty"`

	// phone_number_verified: whether the Provider took affirmative steps to
	// establish that the End-User controlled PhoneNumber at the time it was
	// verified. When this is true, PhoneNumber MUST be in E.164 format. A nil
	// value means the Provider did not send the claim, which is not the same as
	// it sending false.
	PhoneNumberVerified *bool `json:"phone_number_verified,omitempty"`

	// groups: the group memberships of the End-User. This is not an OIDC
	// standard claim; it is a common Provider extension with no agreed-on
	// meaning, so which Providers send it, and what the values name, varies.
	Groups []string `json:"groups,omitempty"`

	// Inline holds the claims that have no field above, such as
	// Provider-specific extensions. The ",inline" option makes json v2 treat it
	// as an inline sink, so a single Unmarshal decodes the known claims and
	// collects the rest here. Each member is kept as raw JSON so its value can
	// be decoded on demand into the type the caller expects, without an eager
	// decode to any that would lose number precision.
	Inline map[string]jsontext.Value `json:",inline"`
}

func DecodeOpenidUserInfo(bytes []byte) (*OpenidUserInfo, error) {
	userInfo := &OpenidUserInfo{}
	err := json.Unmarshal(bytes, userInfo)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return userInfo, nil
}
