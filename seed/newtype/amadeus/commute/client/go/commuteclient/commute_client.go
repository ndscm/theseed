package commuteclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/grpclog"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/proto/commutepb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/proto/commutepbconnect"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

var flagAmadeusCommuteServiceServer = seedflag.DefineString("amadeus_commute_service_server", "http://127.0.0.1:2623", "Amadeus commute service server address")

type AmadeusCommuteClient struct {
	client commutepbconnect.AmadeusCommuteServiceClient
}

func (c *AmadeusCommuteClient) SendBrainInput(
	ctx context.Context,
	input *brainpb.BrainInput,
) error {
	_, err := c.client.SendBrainInput(ctx, connect.NewRequest(
		&commutepb.SendBrainInputRequest{
			BrainInput: input,
		}))
	return err
}

// StartTerminal opens a terminal on the agent's playpen. The caller sends a start
// frame first, then keystrokes and resizes, and receives terminal output until
// the shell exits. The terminal lives as long as the stream.
func (c *AmadeusCommuteClient) StartTerminal(
	ctx context.Context,
) *connect.BidiStreamForClient[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame] {
	return c.client.StartTerminal(ctx)
}

type amadeusCommuteClientOptions struct {
	httpClient *http.Client
	server     string
}

type AmadeusCommuteClientOption func(*amadeusCommuteClientOptions)

func WithHttpClient(httpClient *http.Client) AmadeusCommuteClientOption {
	return func(opts *amadeusCommuteClientOptions) {
		opts.httpClient = httpClient
	}
}

func WithServer(server string) AmadeusCommuteClientOption {
	return func(opts *amadeusCommuteClientOptions) {
		opts.server = server
	}
}

func NewAmadeusCommuteClient(opts ...AmadeusCommuteClientOption) *AmadeusCommuteClient {
	o := &amadeusCommuteClientOptions{
		httpClient: http.DefaultClient,
		server:     flagAmadeusCommuteServiceServer.Get(),
	}
	for _, opt := range opts {
		opt(o)
	}
	client := commutepbconnect.NewAmadeusCommuteServiceClient(
		seedbearer.InterceptBearerTransport(o.httpClient),
		o.server,
		connect.WithInterceptors(grpclog.NewLogInterceptor()),
	)
	return &AmadeusCommuteClient{client}
}
