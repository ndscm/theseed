package service

import (
	"context"
	"sort"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/hooin/onsite"
	"github.com/ndscm/theseed/seed/newtype/hooin/roster/proto/rosterpb"
)

type HooinRosterService struct {
	office *onsite.Office
}

func (svc *HooinRosterService) GetTeam(
	ctx context.Context,
	req *connect.Request[rosterpb.GetTeamRequest],
) (*connect.Response[rosterpb.Team], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	team := &rosterpb.Team{
		TeamUuid:    svc.office.Team.GetTeamUuid(),
		Handle:      svc.office.Team.GetHandle(),
		DisplayName: svc.office.Team.GetDisplayName(),
	}
	return connect.NewResponse(team), nil
}

func (svc *HooinRosterService) ListTeamMembers(
	ctx context.Context,
	req *connect.Request[rosterpb.ListTeamMembersRequest],
) (*connect.Response[rosterpb.ListTeamMembersResponse], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	members := svc.office.Team.ListMembers()
	teamMembers := []*rosterpb.TeamMember{}
	for personId, person := range members {
		teamMembers = append(teamMembers, &rosterpb.TeamMember{
			TeamUuid:    svc.office.Team.GetTeamUuid(),
			PersonId:    personId,
			Handle:      person.GetHandle(),
			DisplayName: person.GetDisplayName(),
			OnDuty:      svc.office.GetDuty(personId) != nil,
		})
	}
	sort.Slice(teamMembers, func(i, j int) bool {
		return teamMembers[i].GetPersonId() < teamMembers[j].GetPersonId()
	})

	resp := &rosterpb.ListTeamMembersResponse{
		TeamMembers: teamMembers,
	}
	return connect.NewResponse(resp), nil
}

func NewHooinRosterService(office *onsite.Office) *HooinRosterService {
	return &HooinRosterService{
		office: office,
	}
}
