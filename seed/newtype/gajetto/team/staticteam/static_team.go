package staticteam

import (
	"context"
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team"
	"google.golang.org/grpc/codes"
)

var flagStaticTeamFile = seedflag.DefineString("static_team_file", "/etc/steins/team.json", "Static team file path")

type StaticPerson struct {
	personId string `json:"-"`

	Handle string `json:"handle"`

	DisplayName string `json:"displayName,omitempty"`
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
	staticTeamFilePath := flagStaticTeamFile.Get()
	if strings.HasPrefix(staticTeamFilePath, "~/") {
		homeDir, err := os.UserHomeDir()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		staticTeamFilePath = filepath.Join(homeDir, staticTeamFilePath[2:])
	}
	team := &StaticTeam{}
	_, err := os.Stat(staticTeamFilePath)
	if err != nil && !errors.Is(err, fs.ErrNotExist) {
		return nil, seederr.Wrap(err)
	}
	if err == nil {
		data, err := os.ReadFile(staticTeamFilePath)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		err = json.Unmarshal(data, &team)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		for id, member := range team.Members {
			member.personId = id
		}
	}
	return team, nil
}
