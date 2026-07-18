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

	teamId, err := svc.office.Team.GetTeamId(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	teamHandle, err := svc.office.Team.GetHandle(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	teamDisplayName, err := svc.office.Team.GetDisplayName(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	team := &rosterpb.Team{
		TeamUuid:    teamId,
		Handle:      teamHandle,
		DisplayName: teamDisplayName,
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

	teamId, err := svc.office.Team.GetTeamId(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	members, err := svc.office.Team.ListMembers(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	teamMembers := []*rosterpb.TeamMember{}
	for personId, person := range members {
		personHandle, err := person.GetHandle(ctx)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		personDisplayName, err := person.GetDisplayName(ctx)
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		personOrganic, err := person.GetOrganic(ctx)
		if err != nil {
			return nil, seederr.Wrap(err)
		}

		teamMembers = append(teamMembers, &rosterpb.TeamMember{
			TeamUuid:    teamId,
			PersonId:    personId,
			Handle:      personHandle,
			DisplayName: personDisplayName,
			Organic:     personOrganic,
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
