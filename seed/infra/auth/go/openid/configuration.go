package openid

import (
	"encoding/json"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type OpenidConfiguration struct {
	// See: https://openid.net/specs/openid-connect-discovery-1_0.html#ProviderMetadata
	Issuer                  string   `json:"issuer"`
	AuthorizationEndpoint   string   `json:"authorization_endpoint"`
	TokenEndpoint           string   `json:"token_endpoint"`
	UserinfoEndpoint        string   `json:"userinfo_endpoint"`
	ScopesSupported         []string `json:"scopes_supported"`
	ResponsesTypesSupported []string `json:"responses_types_supported"`
	GrantTypesSupported     []string `json:"grant_types_supported"`
	SubjectTypesSupported   []string `json:"subject_types_supported"`
	ClaimsSupported         []string `json:"claims_supported"`

	Raw map[string]any `json:"-"`
}

func DecodeOpenidConfiguration(bytes []byte) (*OpenidConfiguration, error) {
	cfg := &OpenidConfiguration{}
	err := json.Unmarshal(bytes, cfg)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	err = json.Unmarshal(bytes, &cfg.Raw)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return cfg, nil
}
