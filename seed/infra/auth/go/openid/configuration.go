package openid

import (
	"encoding/json/jsontext"
	"encoding/json/v2"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type OpenidConfiguration struct {
	// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
	Issuer                 string   `json:"issuer"`
	AuthorizationEndpoint  string   `json:"authorization_endpoint"`
	TokenEndpoint          string   `json:"token_endpoint"`
	UserinfoEndpoint       string   `json:"userinfo_endpoint"`
	JwksUri                string   `json:"jwks_uri"`
	ScopesSupported        []string `json:"scopes_supported"`
	ResponseTypesSupported []string `json:"response_types_supported"`
	GrantTypesSupported    []string `json:"grant_types_supported"`
	SubjectTypesSupported  []string `json:"subject_types_supported"`
	ClaimsSupported        []string `json:"claims_supported"`

	// See: https://datatracker.ietf.org/doc/html/rfc8628#section-7.4
	DeviceAuthorizationEndpoint string `json:"device_authorization_endpoint,omitempty"`

	// Inline holds provider metadata members not matched by a typed field
	// above. The ",inline" option makes json v2 treat it as an inline sink, so
	// a single Unmarshal decodes the known fields and collects the rest here.
	// Each member is kept as raw JSON so its value can be decoded on demand into
	// the type the caller expects, without an eager decode to any that would
	// lose number precision.
	Inline map[string]jsontext.Value `json:",inline"`
}

func DecodeOpenidConfiguration(bytes []byte) (*OpenidConfiguration, error) {
	cfg := &OpenidConfiguration{}
	err := json.Unmarshal(bytes, cfg)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return cfg, nil
}
