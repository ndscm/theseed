package keycloaklogin

import (
	"encoding/json/v2"
	"slices"

	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type ClientResourceAccess struct {
	Roles []string `json:"roles,omitempty"`
}

type ResourceAccess map[string]ClientResourceAccess

func VerifyRole(openidUser *openid.OpenidUserInfo, clientId string, role string) error {
	if clientId == "" {
		clientId = openidUser.Azp
	}
	if clientId == "" {
		return seederr.WrapErrorf("no client id to verify role for user %s (%s)", openidUser.PreferredUsername, openidUser.Sub)
	}
	rawResourceAccess, ok := openidUser.Inline["resource_access"]
	if !ok {
		return seederr.WrapErrorf("no resource access for user %s (%s)", openidUser.PreferredUsername, openidUser.Sub)
	}
	resourceAccess := ResourceAccess{}
	err := json.Unmarshal(rawResourceAccess, &resourceAccess)
	if err != nil {
		return seederr.Wrap(err)
	}
	clientResourceAccess, ok := resourceAccess[clientId]
	if !ok {
		return seederr.WrapErrorf("no resource access to client %s for user %s (%s)", clientId, openidUser.PreferredUsername, openidUser.Sub)
	}
	if !slices.Contains(clientResourceAccess.Roles, role) {
		return seederr.WrapErrorf("no role %s in client %s for user %s (%s)", role, clientId, openidUser.PreferredUsername, openidUser.Sub)
	}
	return nil
}
