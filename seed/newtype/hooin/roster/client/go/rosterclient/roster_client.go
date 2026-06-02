package rosterclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/grpclog"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/newtype/hooin/roster/proto/rosterpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/roster/proto/rosterpbconnect"
)

var flagHooinRosterServiceServer = seedflag.DefineString("hooin_roster_service_server", "http://127.0.0.1:4664", "Hooin roster service server address")

func HooinRosterServiceServer() string {
	return flagHooinRosterServiceServer.Get()
}

type HooinRosterClient struct {
	client rosterpbconnect.HooinRosterServiceClient
}

func (c *HooinRosterClient) GetTeam(
	ctx context.Context,
) (*rosterpb.Team, error) {
	resp, err := c.client.GetTeam(ctx, connect.NewRequest(&rosterpb.GetTeamRequest{}))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return resp.Msg, nil
}

func (c *HooinRosterClient) ListTeamMembers(
	ctx context.Context,
) (*rosterpb.ListTeamMembersResponse, error) {
	resp, err := c.client.ListTeamMembers(ctx, connect.NewRequest(&rosterpb.ListTeamMembersRequest{}))
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return resp.Msg, nil
}

type hooinRosterClientOptions struct {
	httpClient *http.Client
	server     string
}

type HooinRosterClientOption func(*hooinRosterClientOptions)

func WithHttpClient(httpClient *http.Client) HooinRosterClientOption {
	return func(opts *hooinRosterClientOptions) {
		opts.httpClient = httpClient
	}
}

func WithServer(server string) HooinRosterClientOption {
	return func(opts *hooinRosterClientOptions) {
		opts.server = server
	}
}

func NewHooinRosterClient(opts ...HooinRosterClientOption) *HooinRosterClient {
	options := &hooinRosterClientOptions{
		httpClient: http.DefaultClient,
		server:     HooinRosterServiceServer(),
	}
	for _, opt := range opts {
		opt(options)
	}
	client := rosterpbconnect.NewHooinRosterServiceClient(
		seedbearer.InterceptBearerTransport(options.httpClient),
		options.server,
		connect.WithInterceptors(grpclog.NewLogInterceptor()),
	)
	return &HooinRosterClient{client}
}
