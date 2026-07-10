package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
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

// StartTerminal runs a terminal for the life of the stream. It is the only RPC
// here that needs the commute connection to carry a stream in both directions
// at once: the shell's output must reach the caller while the caller is still
// typing.
func (svc *AmadeusCommuteService) StartTerminal(
	ctx context.Context,
	stream *connect.BidiStream[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame],
) error {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization

	err = startTerminal(ctx, svc.conscious, stream)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func NewAmadeusCommuteService(conscious *onduty.Conscious) *AmadeusCommuteService {
	return &AmadeusCommuteService{
		conscious: conscious,
	}
}
