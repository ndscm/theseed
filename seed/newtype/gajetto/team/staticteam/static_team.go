package staticteam

import (
	"encoding/json"
	"errors"
	"io/fs"
	"os"
	"path/filepath"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team"
	"google.golang.org/grpc/codes"
)

var flagStaticTeamFile = seedflag.DefineString("static_team_file", "/etc/gajetto/team.json", "Static team file path")

type StaticPerson struct {
	Handle string `json:"handle"`

	DisplayName string `json:"displayName,omitempty"`

	Token string `json:"token,omitempty"`
}

func (p *StaticPerson) GetHandle() string {
	return p.Handle
}

func (p *StaticPerson) GetDisplayName() string {
	return p.DisplayName
}

func (p *StaticPerson) Auth(token string) error {
	if p.Token == "" || p.Token != token {
		return seederr.CodeErrorf(codes.Unauthenticated, "invalid token")
	}
	return nil
}

type StaticTeam struct {
	Handle string `json:"handle"`

	DisplayName string `json:"displayName,omitempty"`

	Members map[string]*StaticPerson `json:"members"`
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

func (t *StaticTeam) Auth(token string) (personId string, err error) {
	for id, member := range t.Members {
		if member.Auth(token) == nil {
			return id, nil
		}
	}
	return "", seederr.CodeErrorf(codes.Unauthenticated, "invalid token")
}

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
	}
	return team, nil
}
