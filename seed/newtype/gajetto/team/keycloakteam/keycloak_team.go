// Package keycloakteam implements a [team.Team] backed by a Keycloak realm. The
// member roster is loaded from the Keycloak admin REST API rather than a static
// file: the members are the enabled users of the realm. The team authenticates to
// Keycloak with its own service account (client credentials), independent of
// any calling user's bearer token.
package keycloakteam

import (
	"context"
	"sync"

	"github.com/ndscm/theseed/seed/cloud/keycloak/client/go/keycloak"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/auth/go/openid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team"
	"google.golang.org/grpc/codes"
)

type KeycloakPerson struct {
	user *keycloak.UserRepresentation
}

func (p *KeycloakPerson) GetPersonId(ctx context.Context) (string, error) {
	if p.user == nil {
		return "", seederr.CodeErrorf(codes.NotFound, "keycloak user is nil")
	}
	return p.user.Id, nil
}

func (p *KeycloakPerson) GetHandle(ctx context.Context) (string, error) {
	if p.user == nil {
		return "", seederr.CodeErrorf(codes.NotFound, "keycloak user is nil")
	}
	return p.user.Username, nil
}

func (p *KeycloakPerson) GetDisplayName(ctx context.Context) (string, error) {
	if p.user == nil {
		return "", seederr.CodeErrorf(codes.NotFound, "keycloak user is nil")
	}
	return p.user.FirstName, nil
}

func (p *KeycloakPerson) GetOrganic(ctx context.Context) (string, error) {
	if p.user == nil {
		return "", seederr.CodeErrorf(codes.NotFound, "keycloak user is nil")
	}
	organic := "carbon"
	organicAttributes, ok := p.user.Attributes["organic"]
	if ok && len(organicAttributes) > 0 {
		organic = organicAttributes[0]
	}
	return organic, nil
}

var _ team.Person = (*KeycloakPerson)(nil)

type KeycloakTeam struct {
	keycloakClient *keycloak.KeycloakClient

	realmMutex sync.RWMutex
	realm      *keycloak.RealmRepresentation
}

func (t *KeycloakTeam) getRealm() *keycloak.RealmRepresentation {
	t.realmMutex.RLock()
	defer t.realmMutex.RUnlock()
	return t.realm
}

func (t *KeycloakTeam) loadRealm(ctx context.Context) (*keycloak.RealmRepresentation, error) {
	realm := t.getRealm()
	if realm != nil {
		return realm, nil
	}

	t.realmMutex.Lock()
	defer t.realmMutex.Unlock()

	if t.realm != nil {
		return t.realm, nil
	}
	realm, err := t.keycloakClient.GetRealm(ctx, false)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	t.realm = realm
	return t.realm, nil
}

func (t *KeycloakTeam) GetTeamId(ctx context.Context) (string, error) {
	realm, err := t.loadRealm(ctx)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return realm.Id, nil
}

func (t *KeycloakTeam) GetHandle(ctx context.Context) (string, error) {
	realm, err := t.loadRealm(ctx)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return realm.Realm, nil
}

func (t *KeycloakTeam) GetDisplayName(ctx context.Context) (string, error) {
	realm, err := t.loadRealm(ctx)
	if err != nil {
		return "", seederr.Wrap(err)
	}
	return realm.DisplayName, nil
}

func (t *KeycloakTeam) GetMember(ctx context.Context, personId string) (team.Person, error) {
	user, err := t.keycloakClient.GetUser(ctx, true, personId)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return &KeycloakPerson{
		user: user,
	}, nil
}

func (t *KeycloakTeam) GetMemberByHandle(ctx context.Context, personHandle string) (team.Person, error) {
	user, err := t.keycloakClient.GetUserByHandle(ctx, true, personHandle)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return &KeycloakPerson{
		user: user,
	}, nil
}

func (t *KeycloakTeam) ListMembers(ctx context.Context) (map[string]team.Person, error) {
	users, err := t.keycloakClient.ListUsers(ctx, true)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	members := map[string]team.Person{}
	for _, user := range users {
		if !user.Enabled {
			continue
		}
		members[user.Id] = &KeycloakPerson{
			user: &user,
		}
	}
	return members, nil
}

func (t *KeycloakTeam) Auth(ctx context.Context) (string, error) {
	openidUser, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return "", seederr.CodeErrorf(codes.Unauthenticated, "user not logged in")
	}
	return openidUser.Sub, nil
}

var _ team.Team = (*KeycloakTeam)(nil)

func ConnectTeam(openidClient *openid.OpenidClient) (*KeycloakTeam, error) {
	keycloakClient, err := keycloak.CreateKeycloakClient(openidClient)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	keycloakTeam := &KeycloakTeam{
		keycloakClient: keycloakClient,
	}
	return keycloakTeam, nil
}
