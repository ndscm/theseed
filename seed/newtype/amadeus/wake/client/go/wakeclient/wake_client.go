package wakeclient

import (
	"context"
	"net/http"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/flag/go/seedflag"
	"github.com/ndscm/theseed/seed/infra/grpc/go/seedgrpc"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/newtype/amadeus/wake/proto/wakepb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/wake/proto/wakepbconnect"
)

var flagAmadeusServiceServer = seedflag.DefineString("amadeus_service_server", "http://127.0.0.1:2623", "Amadeus service server address")

type AmadeusWakeClient struct {
	client wakepbconnect.AmadeusWakeServiceClient
}

func NewAmadeusWakeClient(server string) *AmadeusWakeClient {
	if server == "" {
		server = flagAmadeusServiceServer.Get()
	}
	client := wakepbconnect.NewAmadeusWakeServiceClient(
		seedbearer.InterceptBearerTransport(http.DefaultClient),
		server,
		connect.WithInterceptors(seedgrpc.NewLogInterceptor()),
	)
	return &AmadeusWakeClient{client}
}

func (c *AmadeusWakeClient) Wake(
	ctx context.Context,
	hooinDirectServer string,
	token string,
) error {
	_, err := c.client.Wake(ctx, connect.NewRequest(&wakepb.WakeRequest{
		HooinDirectServer: hooinDirectServer,
		Token:             token,
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (c *AmadeusWakeClient) RestartBrain(
	ctx context.Context,
	wait bool,
	hotUpgrade bool,
) error {
	_, err := c.client.RestartBrain(ctx, connect.NewRequest(&wakepb.RestartBrainRequest{
		Wait:       wait,
		HotUpgrade: hotUpgrade,
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (c *AmadeusWakeClient) Hibernate(
	ctx context.Context,
	wait bool,
) error {
	_, err := c.client.Hibernate(ctx, connect.NewRequest(&wakepb.HibernateRequest{
		Wait: wait,
	}))
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
