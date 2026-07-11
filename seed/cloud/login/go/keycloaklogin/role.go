package keycloaklogin

import (
	"encoding/json"
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
	rawResourceAccess, ok := openidUser.Raw["resource_access"]
	if !ok {
		return seederr.WrapErrorf("no resource access for user %s (%s)", openidUser.PreferredUsername, openidUser.Sub)
	}
	rawResourceAccessBytes, err := json.Marshal(rawResourceAccess)
	if err != nil {
		return seederr.Wrap(err)
	}
	resourceAccess := ResourceAccess{}
	err = json.Unmarshal(rawResourceAccessBytes, &resourceAccess)
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
