package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepb"
	"github.com/ndscm/theseed/seed/newtype/hooin/onsite"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HooinCommuteService struct {
	office *onsite.Office
}

func (svc *HooinCommuteService) Commute(
	ctx context.Context,
	req *connect.Request[commutepb.CommuteRequest],
	stream *connect.ServerStream[brainpb.BrainInput],
) error {
	return seederr.CodeErrorf(codes.Unimplemented, "Commute is not implemented")
}

func (svc *HooinCommuteService) ReportBrainStep(
	ctx context.Context,
	req *connect.Request[commutepb.ReportBrainStepRequest],
) (*connect.Response[emptypb.Empty], error) {
	personId, err := svc.office.Team.Auth(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	step := req.Msg.GetBrainStep()
	if step == nil {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "brain_step is required")
	}

	svc.office.BroadcastStep(personId, step.GetTopic(), step)

	return connect.NewResponse(&emptypb.Empty{}), nil
}

func NewHooinCommuteService(office *onsite.Office) *HooinCommuteService {
	return &HooinCommuteService{
		office: office,
	}
}
