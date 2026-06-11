package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/proto/commutepb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/onduty"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AmadeusCommuteService struct {
	conscious *onduty.Conscious
}

func (svc *AmadeusCommuteService) SendBrainInput(
	ctx context.Context,
	req *connect.Request[commutepb.SendBrainInputRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization
	brainInput := req.Msg.GetBrainInput()
	if brainInput == nil {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "brain_input is required")
	}
	err = svc.conscious.Input(brainInput)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

func NewAmadeusCommuteService(conscious *onduty.Conscious) *AmadeusCommuteService {
	return &AmadeusCommuteService{
		conscious: conscious,
	}
}
