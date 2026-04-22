package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/amadeus/onduty"
	"github.com/ndscm/theseed/seed/newtype/amadeus/wake/proto/wakepb"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

type AmadeusWakeService struct {
	conscious *onduty.Conscious
}

func (svc *AmadeusWakeService) Initialize(conscious *onduty.Conscious) error {
	svc.conscious = conscious
	return nil
}

func (svc *AmadeusWakeService) Wake(
	ctx context.Context,
	req *connect.Request[wakepb.WakeRequest],
) (*connect.Response[emptypb.Empty], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization
	err = svc.conscious.Wake(ctx, req.Msg.GetToken(), req.Msg.GetHooinDirectServer())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
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
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization
	done := svc.conscious.Hibernate()
	if done == nil {
		return connect.NewResponse(&emptypb.Empty{}), nil
	}

	if req.Msg.GetWait() {
		select {
		case <-done:
		case <-ctx.Done():
			return nil, seederr.Wrap(ctx.Err())
		}
	}

	return connect.NewResponse(&emptypb.Empty{}), nil
}
