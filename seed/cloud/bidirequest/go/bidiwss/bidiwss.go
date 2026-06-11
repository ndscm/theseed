package bidiwss

import (
	"context"

	"github.com/coder/websocket"
	"github.com/ndscm/theseed/seed/cloud/bidirequest/proto/bidirequestpb"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"google.golang.org/protobuf/proto"
)

// readLimit caps a single websocket frame. Payloads carry whole dumped
// http requests and responses, which can exceed the library default read
// limit of 32KiB, but an unlimited frame would let a misbehaving peer
// exhaust memory.
const readLimit = 64 * 1024 * 1024

// BidiWebSocketStream carries Payload messages as binary websocket frames.
// Its context is canceled when the stream is closed or broken, so peers
// can watch Context().Done() for the end of the connection.
type BidiWebSocketStream struct {
	ctx    context.Context
	cancel context.CancelFunc
	conn   *websocket.Conn
}

func (s *BidiWebSocketStream) Context() context.Context {
	return s.ctx
}

func (s *BidiWebSocketStream) Send(payload *bidirequestpb.Payload) error {
	data, err := proto.Marshal(payload)
	if err != nil {
		return seederr.Wrap(err)
	}
	err = s.conn.Write(s.ctx, websocket.MessageBinary, data)
	if err != nil {
		s.cancel()
		return seederr.Wrap(err)
	}
	return nil
}

func (s *BidiWebSocketStream) Receive() (*bidirequestpb.Payload, error) {
	msgType, data, err := s.conn.Read(s.ctx)
	if err != nil {
		s.cancel()
		return nil, seederr.Wrap(err)
	}
	if msgType != websocket.MessageBinary {
		s.cancel()
		return nil, seederr.WrapErrorf("unexpected websocket message type: %v", msgType)
	}
	payload := &bidirequestpb.Payload{}
	err = proto.Unmarshal(data, payload)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return payload, nil
}

func (s *BidiWebSocketStream) Close() error {
	s.cancel()
	err := s.conn.Close(websocket.StatusNormalClosure, "")
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func NewBidiWebSocketStream(ctx context.Context, conn *websocket.Conn) *BidiWebSocketStream {
	conn.SetReadLimit(readLimit)
	ctx, cancel := context.WithCancel(ctx)
	return &BidiWebSocketStream{
		ctx:    ctx,
		cancel: cancel,
		conn:   conn,
	}
}
