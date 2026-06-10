package bidirequestclient

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidiwss"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/http/go/seedbearer"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type BidirequestClient struct {
	server     string
	httpClient *http.Client
}

func (c *BidirequestClient) Connect(
	ctx context.Context,
) (*bidiwss.BidiWebSocketStream, error) {
	seedlog.Infof("Bidirequest connecting. server=%v", c.server)
	conn, _, err := websocket.Dial(ctx,
		c.server+"/seed.cloud.bidirequest.proto.BidirequestService/Connect",
		&websocket.DialOptions{
			HTTPClient: c.httpClient,
		})
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return bidiwss.NewBidiWebSocketStream(ctx, conn), nil
}

func NewBidirequestClient(server string) *BidirequestClient {
	return &BidirequestClient{
		server:     server,
		httpClient: seedbearer.InterceptBearerTransport(http.DefaultClient),
	}
}
