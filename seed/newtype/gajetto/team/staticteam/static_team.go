package staticteam

import (
	"context"
	"encoding/json"

	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team"
	"google.golang.org/grpc/codes"
)

var flagStaticTeamFile = seedflag.DefineFile(
	"static_team_file", "/etc/steins/team.json",
	"Static team file path",
)

type StaticPerson struct {
	personId string `json:"-"`

	Handle string `json:"handle"`

	DisplayName string `json:"displayName,omitempty"`

	Organic string `json:"organic,omitempty"`
}

func (p *StaticPerson) GetPersonId() string {
	return p.personId
}

func (p *StaticPerson) GetHandle() string {
	return p.Handle
}

func (p *StaticPerson) GetDisplayName() string {
	return p.DisplayName
}

func (p *StaticPerson) GetOrganic() string {
	return p.Organic
}

var _ team.Person = (*StaticPerson)(nil)

type StaticTeam struct {
	Handle string `json:"handle"`

	DisplayName string `json:"displayName,omitempty"`

	Members map[string]*StaticPerson `json:"members"`
}

func (t *StaticTeam) GetTeamUuid() string {
	return ""
}

func (t *StaticTeam) GetHandle() string {
	return t.Handle
}

func (t *StaticTeam) GetDisplayName() string {
	return t.DisplayName
}

func (t *StaticTeam) GetMember(personId string) (team.Person, bool) {
	member, ok := t.Members[personId]
	if !ok {
		return nil, false
	}
	return member, true
}

func (t *StaticTeam) ListMembers() map[string]team.Person {
	members := map[string]team.Person{}
	for id, p := range t.Members {
		members[id] = p
	}
	return members
}

func (t *StaticTeam) Auth(ctx context.Context) (personId string, err error) {
	openidUser, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return "", seederr.CodeErrorf(codes.Unauthenticated, "user not logged in")
	}
	member, ok := t.Members[openidUser.PreferredUsername]
	if !ok {
		return "", seederr.CodeErrorf(codes.PermissionDenied, "permission denied")
	}
	return member.personId, nil
}

var _ team.Team = (*StaticTeam)(nil)

func LoadTeam() (team.Team, error) {
	staticTeamBytes, err := flagStaticTeamFile.Load()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	team := &StaticTeam{}
	if len(staticTeamBytes) > 0 {
		err = json.Unmarshal(staticTeamBytes, &team)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		for id, member := range team.Members {
			member.personId = id
		}
	}
	return team, nil
}
