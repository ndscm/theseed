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

func (svc *HooinCommuteService) Initialize(office *onsite.Office) error {
	svc.office = office
	return nil
}

func (svc *HooinCommuteService) Commute(
	ctx context.Context,
	req *connect.Request[commutepb.CommuteRequest],
	stream *connect.ServerStream[brainpb.BrainInput],
) error {
	personId, err := svc.office.Team.Auth(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}

	duty := onsite.NewPersonDuty(stream)
	err = svc.office.SetDuty(personId, duty)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer svc.office.ClearDuty(personId)

	// Flush response headers immediately so the client's Commute call
	// returns once the session is established, instead of blocking until
	// the first BrainInput is forwarded.
	err = duty.Flush()
	if err != nil {
		return seederr.Wrap(err)
	}

	<-ctx.Done()
	return nil
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
