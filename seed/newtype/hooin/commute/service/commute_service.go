package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type HooinCommuteService struct {
}

func (svc *HooinCommuteService) Initialize() error {
	return nil
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
	return nil, seederr.CodeErrorf(codes.Unimplemented, "ReportBrainStep is not implemented")
}
