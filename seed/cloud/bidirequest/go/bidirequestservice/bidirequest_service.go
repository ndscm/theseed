package bidirequestservice

import (
	"context"
	"net/http"

	"github.com/coder/websocket"
	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidirequest"
	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidiwss"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type ConnectHandler interface {
	HandleConnect(ctx context.Context, stream bidirequest.PayloadStream) error
}

type BidirequestServiceHandler struct {
	connectHandler ConnectHandler
}

func (svc *BidirequestServiceHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	seedlog.Debugf("Received new connection")
	conn, err := websocket.Accept(w, r, nil)
	if err != nil {
		seedlog.Errorf("Websocket accept failed: %v", err)
		return
	}
	stream := bidiwss.NewBidiWebSocketStream(r.Context(), conn)
	defer stream.Close()
	err = svc.connectHandler.HandleConnect(stream.Context(), stream)
	if err != nil {
		seedlog.Errorf("Connect handler failed: %v", err)
	}
}

var _ http.Handler = (*BidirequestServiceHandler)(nil)

func NewBidirequestServiceHandler(connectHandler ConnectHandler) (string, *BidirequestServiceHandler) {
	svc := &BidirequestServiceHandler{
		connectHandler: connectHandler,
	}
	return "/seed.cloud.bidirequest.proto.BidirequestService/Connect", svc
}
