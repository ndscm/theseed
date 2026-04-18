package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/amadeus/wake/proto/wakepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AmadeusWakeService struct {
}

func (svc *AmadeusWakeService) Initialize() error {
	return nil
}

func (svc *AmadeusWakeService) Wake(
	ctx context.Context,
	req *connect.Request[wakepb.WakeRequest],
) (*connect.Response[emptypb.Empty], error) {
	return nil, seederr.CodeErrorf(codes.Unimplemented, "Wake is not implemented")
}

func (svc *AmadeusWakeService) RestartBrain(
	ctx context.Context,
	req *connect.Request[wakepb.RestartBrainRequest],
) (*connect.Response[emptypb.Empty], error) {
	return nil, seederr.CodeErrorf(codes.Unimplemented, "RestartBrain is not implemented")
}

func (svc *AmadeusWakeService) Hibernate(
	ctx context.Context,
	req *connect.Request[wakepb.HibernateRequest],
) (*connect.Response[emptypb.Empty], error) {
	return nil, seederr.CodeErrorf(codes.Unimplemented, "Hibernate is not implemented")
}
