package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepb"
	"google.golang.org/grpc/codes"
)

type HooinDictateService struct {
}

func (svc *HooinDictateService) Initialize() error {
	return nil
}

func (svc *HooinDictateService) SendBrainInput(
	ctx context.Context,
	req *connect.Request[dictatepb.SendBrainInputRequest],
) (*connect.Response[brainpb.BrainStep], error) {
	return nil, seederr.CodeErrorf(codes.Unimplemented, "SendBrainInput is not implemented")
}

func (svc *HooinDictateService) SendBrainInputStreamBrainStep(
	ctx context.Context,
	req *connect.Request[dictatepb.SendBrainInputRequest],
	stream *connect.ServerStream[brainpb.BrainStep],
) error {
	return seederr.CodeErrorf(codes.Unimplemented, "SendBrainInputStreamBrainStep is not implemented")
}

func (svc *HooinDictateService) SubscribeBrainStep(
	ctx context.Context,
	req *connect.Request[dictatepb.SubscribeBrainStepRequest],
	stream *connect.ServerStream[brainpb.BrainStep],
) error {
	return seederr.CodeErrorf(codes.Unimplemented, "SubscribeBrainStep is not implemented")
}
