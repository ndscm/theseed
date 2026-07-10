package invadeclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/grpclog"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/invade/proto/invadepb"
	"github.com/ndscm/theseed/seed/newtype/hooin/invade/proto/invadepbconnect"
)

var flagHooinInvadeServiceServer = seedflag.DefineString("hooin_invade_service_server", "http://127.0.0.1:4664", "Hooin invade service server address")

func HooinInvadeServiceServerFlag() string {
	return flagHooinInvadeServiceServer.Get()
}

type HooinInvadeClient struct {
	client invadepbconnect.HooinInvadeServiceClient
}

// StartTerminal opens a terminal on a person's workstation and streams what it
// prints, until the shell exits or the caller hangs up.
// See HooinInvadeService.StartTerminal.
func (c *HooinInvadeClient) StartTerminal(
	ctx context.Context,
	personId string,
	start *terminalpb.TerminalInputFrame,
) (*connect.ServerStreamForClient[terminalpb.TerminalOutputFrame], error) {
	return c.client.StartTerminal(ctx, connect.NewRequest(&invadepb.StartTerminalRequest{
		PersonId: personId,
		Start:    start,
	}))
}

// SendTerminalInput types at a terminal StartTerminal opened, or tells it its
// window has been resized. personId is the person whose workstation the
// terminal runs on — the one it was opened with — and the frame's session uuid
// names which of that person's terminals.
//
// A caller with more than one keystroke to deliver must await each call before
// making the next, or the shell may read them out of order.
// See HooinInvadeService.SendTerminalInput.
func (c *HooinInvadeClient) SendTerminalInput(
	ctx context.Context,
	personId string,
	frame *terminalpb.TerminalInputFrame,
) error {
	_, err := c.client.SendTerminalInput(ctx, connect.NewRequest(&invadepb.SendTerminalInputRequest{
		PersonId: personId,
		Frame:    frame,
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

type hooinInvadeClientOptions struct {
	httpClient *http.Client
	server     string
}

type HooinInvadeClientOption func(*hooinInvadeClientOptions)

func WithHttpClient(httpClient *http.Client) HooinInvadeClientOption {
	return func(opts *hooinInvadeClientOptions) {
		opts.httpClient = httpClient
	}
}

func WithServer(server string) HooinInvadeClientOption {
	return func(opts *hooinInvadeClientOptions) {
		opts.server = server
	}
}

func NewHooinInvadeClient(opts ...HooinInvadeClientOption) *HooinInvadeClient {
	o := &hooinInvadeClientOptions{
		httpClient: http.DefaultClient,
		server:     flagHooinInvadeServiceServer.Get(),
	}
	for _, opt := range opts {
		opt(o)
	}
	client := invadepbconnect.NewHooinInvadeServiceClient(
		seedbearer.InterceptBearerTransport(o.httpClient),
		o.server,
		connect.WithInterceptors(grpclog.NewLogInterceptor()),
	)
	return &HooinInvadeClient{client}
}
