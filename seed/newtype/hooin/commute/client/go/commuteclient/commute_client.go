package commuteclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepbconnect"
)

var flagHooinCommuteServiceServer = seedflag.DefineString("hooin_commute_service_server", "http://127.0.0.1:4664", "Hooin commute service server address")

type HooinCommuteClient struct {
	client commutepbconnect.HooinCommuteServiceClient
}

func NewHooinCommuteClient(server string) *HooinCommuteClient {
	if server == "" {
		server = flagHooinCommuteServiceServer.Get()
	}
	client := commutepbconnect.NewHooinCommuteServiceClient(
		http.DefaultClient,
		server,
		connect.WithInterceptors(seedgrpc.NewLogInterceptor()),
	)
	return &HooinCommuteClient{client}
}

// Commute opens the Amadeus-side stream that delivers BrainInputs to the
// commuting agent. See HooinCommuteService.Commute.
func (c *HooinCommuteClient) Commute(
	ctx context.Context,
	token string,
) (*connect.ServerStreamForClient[brainpb.BrainInput], error) {
	return c.client.Commute(ctx, connect.NewRequest(&commutepb.CommuteRequest{
		Token: token,
	}))
}

// ReportBrainStep publishes a BrainStep emitted by the commuting agent.
// See HooinCommuteService.ReportBrainStep.
func (c *HooinCommuteClient) ReportBrainStep(
	ctx context.Context,
	token string,
	step *brainpb.BrainStep,
) error {
	_, err := c.client.ReportBrainStep(ctx, connect.NewRequest(&commutepb.ReportBrainStepRequest{
		Token:     token,
		BrainStep: step,
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
