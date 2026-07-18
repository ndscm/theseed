package service

import (
	"context"
	"sort"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team"
	"github.com/ndscm/theseed/seed/newtype/hooin/onsite"
	"github.com/ndscm/theseed/seed/newtype/hooin/roster/proto/rosterpb"
	"google.golang.org/grpc/codes"
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

// GetTeamMember looks up a single member, named by either person_id or handle.
// person_id wins when both are given; naming neither is an error.
func (svc *HooinRosterService) GetTeamMember(
	ctx context.Context,
	req *connect.Request[rosterpb.GetTeamMemberRequest],
) (*connect.Response[rosterpb.TeamMember], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	teamId, err := svc.office.Team.GetTeamId(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	person := (team.Person)(nil)
	switch {
	case req.Msg.GetPersonId() != "":
		person, err = svc.office.Team.GetMember(ctx, req.Msg.GetPersonId())
	case req.Msg.GetHandle() != "":
		person, err = svc.office.Team.GetMemberByHandle(ctx, req.Msg.GetHandle())
	default:
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "person_id or handle is required")
	}
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// The lookup may have been by handle, so the id that keys the duty is
	// whatever the person reports, not the request field.
	personId, err := person.GetPersonId(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
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

	teamMember := &rosterpb.TeamMember{
		TeamUuid:    teamId,
		PersonId:    personId,
		Handle:      personHandle,
		DisplayName: personDisplayName,
		Organic:     personOrganic,
		OnDuty:      svc.office.GetDuty(personId) != nil,
	}
	return connect.NewResponse(teamMember), nil
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
