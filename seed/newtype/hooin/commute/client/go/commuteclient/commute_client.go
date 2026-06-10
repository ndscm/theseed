package commuteclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/grpclog"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/proto/commutepbconnect"
)

var flagHooinCommuteServiceServer = seedflag.DefineString("hooin_commute_service_server", "http://127.0.0.1:4664", "Hooin commute service server address")

func HooinCommuteServiceServerFlag() string {
	return flagHooinCommuteServiceServer.Get()
}

type HooinCommuteClient struct {
	client commutepbconnect.HooinCommuteServiceClient
}

type hooinCommuteClientOptions struct {
	httpClient *http.Client
	server     string
}

type HooinCommuteClientOption func(*hooinCommuteClientOptions)

func WithHttpClient(httpClient *http.Client) HooinCommuteClientOption {
	return func(opts *hooinCommuteClientOptions) {
		opts.httpClient = httpClient
	}
}

func WithServer(server string) HooinCommuteClientOption {
	return func(opts *hooinCommuteClientOptions) {
		opts.server = server
	}
}

func NewHooinCommuteClient(opts ...HooinCommuteClientOption) *HooinCommuteClient {
	o := &hooinCommuteClientOptions{
		httpClient: http.DefaultClient,
		server:     flagHooinCommuteServiceServer.Get(),
	}
	for _, opt := range opts {
		opt(o)
	}
	client := commutepbconnect.NewHooinCommuteServiceClient(
		seedbearer.InterceptBearerTransport(o.httpClient),
		o.server,
		connect.WithInterceptors(grpclog.NewLogInterceptor()),
	)
	return &HooinCommuteClient{client}
}

// Commute opens the Amadeus-side stream that delivers BrainInputs to the
// commuting agent. See HooinCommuteService.Commute.
func (c *HooinCommuteClient) Commute(
	ctx context.Context,
) (*connect.ServerStreamForClient[brainpb.BrainInput], error) {
	return c.client.Commute(ctx, connect.NewRequest(&commutepb.CommuteRequest{}))
}

// ReportBrainStep publishes a BrainStep emitted by the commuting agent.
// See HooinCommuteService.ReportBrainStep.
func (c *HooinCommuteClient) ReportBrainStep(
	ctx context.Context,
	step *brainpb.BrainStep,
) error {
	_, err := c.client.ReportBrainStep(ctx, connect.NewRequest(&commutepb.ReportBrainStepRequest{
		BrainStep: step,
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
