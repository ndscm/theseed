package dictateclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepb"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepbconnect"
)

var flagHooinDictateServiceServer = seedflag.DefineString("hooin_dictate_service_server", "http://127.0.0.1:4664", "Hooin dictate service server address")

type HooinDictateClient struct {
	client dictatepbconnect.HooinDictateServiceClient
}

func NewHooinDictateClient(server string) *HooinDictateClient {
	if server == "" {
		server = flagHooinDictateServiceServer.Get()
	}
	client := dictatepbconnect.NewHooinDictateServiceClient(
		seedbearer.InterceptBearerTransport(http.DefaultClient),
		server,
		connect.WithInterceptors(seedgrpc.NewLogInterceptor()),
	)
	return &HooinDictateClient{client}
}

// SendBrainInput delivers a BrainInput and returns the terminal `result`
// BrainStep once processing completes. Intermediate steps are dropped.
// See HooinDictateService.SendBrainInput.
func (c *HooinDictateClient) SendBrainInput(
	ctx context.Context,
	personId string,
	input *brainpb.BrainInput,
) (*brainpb.BrainStep, error) {
	resp, err := c.client.SendBrainInput(ctx, connect.NewRequest(&dictatepb.SendBrainInputRequest{
		PersonId:   personId,
		BrainInput: input,
	}))
	if err != nil {
		return nil, err
	}
	return resp.Msg, nil
}

// SendBrainInputStreamBrainStep delivers a BrainInput and streams every
// BrainStep emitted while it is processed.
// See HooinDictateService.SendBrainInputStreamBrainStep.
func (c *HooinDictateClient) SendBrainInputStreamBrainStep(
	ctx context.Context,
	personId string,
	input *brainpb.BrainInput,
) (*connect.ServerStreamForClient[brainpb.BrainStep], error) {
	return c.client.SendBrainInputStreamBrainStep(ctx, connect.NewRequest(&dictatepb.SendBrainInputRequest{
		PersonId:   personId,
		BrainInput: input,
	}))
}

// SubscribeBrainStep opens a long-lived stream of BrainSteps matching the
// given (person, topic) pairs. See HooinDictateService.SubscribeBrainStep.
func (c *HooinDictateClient) SubscribeBrainStep(
	ctx context.Context,
	personTopics []*dictatepb.PersonTopic,
) (*connect.ServerStreamForClient[brainpb.BrainStep], error) {
	return c.client.SubscribeBrainStep(ctx, connect.NewRequest(&dictatepb.SubscribeBrainStepRequest{
		PersonTopics: personTopics,
	}))
}
