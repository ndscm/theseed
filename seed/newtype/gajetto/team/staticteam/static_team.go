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

func (p *StaticPerson) GetPersonId(ctx context.Context) (string, error) {
	return p.personId, nil
}

func (p *StaticPerson) GetHandle(ctx context.Context) (string, error) {
	return p.Handle, nil
}

func (p *StaticPerson) GetDisplayName(ctx context.Context) (string, error) {
	return p.DisplayName, nil
}

func (p *StaticPerson) GetOrganic(ctx context.Context) (string, error) {
	return p.Organic, nil
}

var _ team.Person = (*StaticPerson)(nil)

type StaticTeam struct {
	Handle string `json:"handle"`

	DisplayName string `json:"displayName,omitempty"`

	Members map[string]*StaticPerson `json:"members"`
}

func (t *StaticTeam) GetTeamId(ctx context.Context) (string, error) {
	return "", nil
}

func (t *StaticTeam) GetHandle(ctx context.Context) (string, error) {
	return t.Handle, nil
}

func (t *StaticTeam) GetDisplayName(ctx context.Context) (string, error) {
	return t.DisplayName, nil
}

func (t *StaticTeam) GetMember(ctx context.Context, personId string) (team.Person, error) {
	member, ok := t.Members[personId]
	if !ok {
		return nil, seederr.CodeErrorf(codes.NotFound, "member not found")
	}
	return member, nil
}

func (t *StaticTeam) GetMemberByHandle(ctx context.Context, personHandle string) (team.Person, error) {
	for _, member := range t.Members {
		if member.Handle == personHandle {
			return member, nil
		}
	}
	return nil, seederr.CodeErrorf(codes.NotFound, "member not found")
}

func (t *StaticTeam) ListMembers(ctx context.Context) (map[string]team.Person, error) {
	members := map[string]team.Person{}
	for id, p := range t.Members {
		members[id] = p
	}
	return members, nil
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
