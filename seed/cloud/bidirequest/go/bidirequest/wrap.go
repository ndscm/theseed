package bidirequest

import (
	"bufio"
	"bytes"
	"net/http"
	"net/http/httptest"
	"net/http/httputil"
	"sync"
	"sync/atomic"

	"github.com/ndscm/theseed/seed/cloud/bidirequest/proto/bidirequestpb"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

type PayloadStream interface {
	Send(*bidirequestpb.Payload) error
	Receive() (*bidirequestpb.Payload, error)
	Close() error
}

const SERVER_SIDE = 0
const CLIENT_SIDE = 1

type MuxTransport struct {
	stream PayloadStream

	handler http.Handler

	side   uint32
	nextId atomic.Uint32

	sendMu sync.Mutex

	pendingMu sync.Mutex
	pending   map[uint32]chan *bidirequestpb.Payload

	recvErr error
}

func (m *MuxTransport) handleRequest(payload *bidirequestpb.Payload) {
	req, err := http.ReadRequest(bufio.NewReader(bytes.NewReader(payload.Data)))
	if err != nil {
		return
	}
	rec := httptest.NewRecorder()
	m.handler.ServeHTTP(rec, req)
	respBytes, err := httputil.DumpResponse(rec.Result(), true)
	if err != nil {
		return
	}

	m.sendMu.Lock()
	err = m.stream.Send(&bidirequestpb.Payload{
		StreamId: payload.StreamId,
		Data:     respBytes,
	})
	m.sendMu.Unlock()
	if err != nil {
		seedlog.Warnf("dropped response for stream %d: %v", payload.StreamId, err)
	}
}

func (m *MuxTransport) startReceiveLoop() {
	go func() {
		for {
			payload, err := m.stream.Receive()
			if err != nil {
				m.pendingMu.Lock()
				m.recvErr = err
				for _, ch := range m.pending {
					close(ch)
				}
				m.pending = nil
				m.pendingMu.Unlock()
				return
			}
			if payload.StreamId&1 == m.side {
				m.pendingMu.Lock()
				ch, ok := m.pending[payload.StreamId]
				if ok {
					ch <- payload
					delete(m.pending, payload.StreamId)
				}
				m.pendingMu.Unlock()
			} else {
				go m.handleRequest(payload)
			}
		}
	}()
}

func (m *MuxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	reqBytes, err := httputil.DumpRequestOut(req, true)
	if err != nil {
		return nil, err
	}

	streamId := m.nextId.Add(2) - 2
	ch := make(chan *bidirequestpb.Payload, 1)

	m.pendingMu.Lock()
	if m.pending == nil {
		m.pendingMu.Unlock()
		return nil, m.recvErr
	}
	m.pending[streamId] = ch
	m.pendingMu.Unlock()

	m.sendMu.Lock()
	sendErr := m.stream.Send(&bidirequestpb.Payload{
		StreamId: streamId,
		Data:     reqBytes,
	})
	m.sendMu.Unlock()
	if sendErr != nil {
		m.pendingMu.Lock()
		delete(m.pending, streamId)
		m.pendingMu.Unlock()
		return nil, sendErr
	}

	select {
	case resp, ok := <-ch:
		if !ok {
			return nil, m.recvErr
		}
		return http.ReadResponse(bufio.NewReader(bytes.NewReader(resp.Data)), req)
	case <-req.Context().Done():
		m.pendingMu.Lock()
		delete(m.pending, streamId)
		m.pendingMu.Unlock()
		return nil, req.Context().Err()
	}
}

// Server-initiated streams use even IDs (2, 4, 6, ...).
// Client-initiated streams use odd IDs (1, 3, 5, ...),
func wrapClient(stream PayloadStream, handler http.Handler, side uint32) *http.Client {
	transport := &MuxTransport{
		stream:  stream,
		handler: handler,
		side:    side,
		pending: map[uint32]chan *bidirequestpb.Payload{},
	}
	initialId := uint32(side)
	if initialId == 0 {
		initialId += 2
	}
	transport.nextId.Store(initialId)
	transport.startReceiveLoop()
	return &http.Client{Transport: transport}
}

func WrapClientSide(stream PayloadStream, handler http.Handler) *http.Client {
	return wrapClient(stream, handler, CLIENT_SIDE)
}

func WrapServerSide(stream PayloadStream, handler http.Handler) *http.Client {
	return wrapClient(stream, handler, SERVER_SIDE)
}
