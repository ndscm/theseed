// Package bidirequest multiplexes HTTP request/response pairs, in both
// directions, over a single stream of Payload frames.
//
// The frames ride a WebSocket, which is a plain HTTP/1.1 upgrade and so
// survives the proxies between the two peers (see the README). Nothing in this
// package puts HTTP/2 on the wire.
package bidirequest

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"net/http"
	"sync/atomic"

	"github.com/ndscm/theseed/seed/cloud/bidirequest/proto/bidirequestpb"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

// PayloadStream is the wire the frames ride on: a full-duplex, ordered stream
// of Payload messages to the peer, in practice a WebSocket connection. Both
// sides of a connection wrap one, and the two halves of it are independent —
// Send and Receive may be called concurrently.
type PayloadStream interface {
	Send(*bidirequestpb.Payload) error
	Receive() (*bidirequestpb.Payload, error)
	Close() error
}

// SERVER_SIDE and CLIENT_SIDE say which end of the connection a MuxTransport
// is: the side that accepted the WebSocket, or the side that dialed it. This is
// what keeps the stream ids the two ends open from colliding — the server side
// opens even ids (2, 4, 6, ...) and the client side odd ones (1, 3, 5, ...).
// After the handshake the two sides are symmetric: either may initiate a
// stream.
const SERVER_SIDE = 0
const CLIENT_SIDE = 1

// frameResponseWriter is the http.ResponseWriter handed to a handler serving a
// tunneled request. Every Write becomes a frame, so a streaming handler's
// output reaches the peer as it is produced.
type frameResponseWriter struct {
	conn     *MuxConnection
	streamId uint32

	header      http.Header
	wroteHeader bool
	writeErr    error
}

func (w *frameResponseWriter) Header() http.Header {
	return w.header
}

func (w *frameResponseWriter) WriteHeader(statusCode int) {
	if w.wroteHeader {
		return
	}
	w.wroteHeader = true

	head := &bytes.Buffer{}
	fmt.Fprintf(head, "HTTP/1.1 %d %s\r\n", statusCode, http.StatusText(statusCode))
	err := w.header.WriteSubset(head, responseHeaderExcludes)
	if err != nil {
		w.writeErr = err
		return
	}
	head.WriteString("Transfer-Encoding: chunked\r\n\r\n")

	w.writeErr = w.conn.send(&bidirequestpb.Payload{
		StreamId: w.streamId,
		Data:     head.Bytes(),
	})
}

func (w *frameResponseWriter) Write(data []byte) (int, error) {
	w.WriteHeader(http.StatusOK)
	if w.writeErr != nil {
		return 0, w.writeErr
	}
	if len(data) == 0 {
		return 0, nil
	}
	err := w.conn.sendChunk(w.streamId, data)
	if err != nil {
		w.writeErr = err
		return 0, err
	}
	return len(data), nil
}

// Flush exists so handlers that stream — connect checks for http.Flusher —
// know their output is going out. Every Write is already its own frame, so
// there is nothing held back to flush.
func (w *frameResponseWriter) Flush() {
	w.WriteHeader(http.StatusOK)
}

func (w *frameResponseWriter) finish() {
	w.WriteHeader(http.StatusOK)
	if w.writeErr != nil {
		w.conn.sendReset(w.streamId)
		return
	}
	err := w.conn.sendLastChunk(w.streamId)
	if err != nil {
		seedlog.Warnf("bidirequest: failed to end response on stream %d: %v", w.streamId, err)
	}
}

var _ http.ResponseWriter = (*frameResponseWriter)(nil)
var _ http.Flusher = (*frameResponseWriter)(nil)

// serverRequestBody keeps a handler's Close from draining the request body.
//
// The body net/http hands back from ReadRequest reads to EOF when closed, so
// that a real connection can be reused for the next request. There is no
// connection to reuse here, and the peer may well have stopped sending — a
// handler that fails before reading its input is the ordinary case — so that
// drain would block forever. Closing simply abandons the stream instead.
type serverRequestBody struct {
	io.ReadCloser

	inbound *inboundStream
}

func (b *serverRequestBody) Close() error {
	return b.inbound.Close()
}

// responseBody ties the lifetime of a stream to the response body read off it,
// so a caller that stops reading early — or never reads at all — releases the
// stream and tells the peer to stop sending.
type responseBody struct {
	io.ReadCloser

	conn     *MuxConnection
	streamId uint32
	inbound  *inboundStream
}

// Close abandons the response rather than delegating to the body net/http gave
// us, which reads to EOF so a real connection can be reused. There is no
// connection to reuse, and a caller closing a response the peer is still
// writing — a bidi stream torn down from this side — would wait on that drain
// forever. See serverRequestBody, which has the same problem on the other end.
func (b *responseBody) Close() error {
	// Abandoning a response the peer is still writing means telling it to stop;
	// one it already finished needs nothing but the entry cleaned up.
	if !b.inbound.ended.Load() {
		b.conn.sendReset(b.streamId)
	}
	b.conn.takeStream(b.streamId)
	b.inbound.Close()
	return nil
}

// MuxTransport is the HTTP-facing half: it serves peer-initiated streams
// against handler and dials new ones as an http.RoundTripper, all over conn.
type MuxTransport struct {
	conn *MuxConnection

	handler http.Handler

	side   uint32
	nextId atomic.Uint32
}

// requestHeaderExcludes are the headers written from the request's own fields,
// or dictated by this transport's framing, rather than copied from the header
// map.
var requestHeaderExcludes = map[string]bool{
	"Host":              true,
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Connection":        true,
}

// responseHeaderExcludes are the same for a response: written from the
// response's own fields, or dictated by this transport's framing.
var responseHeaderExcludes = map[string]bool{
	"Content-Length":    true,
	"Transfer-Encoding": true,
	"Connection":        true,
}

// writeRequest serializes req onto the stream as an HTTP/1.1 message, framing
// the body as it is produced rather than buffering it. It returns once the body
// is exhausted, having ended this side's direction of the stream; the response
// arrives independently.
func (m *MuxTransport) writeRequest(streamId uint32, req *http.Request) error {
	if req.Body != nil {
		defer req.Body.Close()
	}

	host := req.Host
	if host == "" {
		host = req.URL.Host
	}
	method := req.Method
	if method == "" {
		method = http.MethodGet
	}

	// A body of unknown length is the normal case here: a streaming client
	// hands over an open pipe, and even a unary connect call leaves
	// ContentLength at zero. So the body is always chunked when there is one,
	// which is also what lets the peer read it incrementally.
	hasBody := req.Body != nil && req.Body != http.NoBody

	head := &bytes.Buffer{}
	fmt.Fprintf(head, "%s %s HTTP/1.1\r\n", method, req.URL.RequestURI())
	fmt.Fprintf(head, "Host: %s\r\n", host)
	if hasBody {
		head.WriteString("Transfer-Encoding: chunked\r\n")
	}
	err := req.Header.WriteSubset(head, requestHeaderExcludes)
	if err != nil {
		return err
	}
	head.WriteString("\r\n")

	err = m.conn.send(&bidirequestpb.Payload{StreamId: streamId, Data: head.Bytes(), End: !hasBody})
	if err != nil {
		return err
	}
	if !hasBody {
		return nil
	}

	buf := make([]byte, bodyChunkSize)
	for {
		n, readErr := req.Body.Read(buf)
		if n > 0 {
			err := m.conn.sendChunk(streamId, buf[:n])
			if err != nil {
				return err
			}
		}
		if readErr == io.EOF {
			break
		}
		if readErr != nil {
			m.conn.sendReset(streamId)
			return readErr
		}
	}
	return m.conn.sendLastChunk(streamId)
}

// serve reads a request off a peer-initiated stream and runs it against the
// local handler, writing the response back on the same stream.
func (m *MuxTransport) serve(streamId uint32, inbound *inboundStream) {
	// serve owns the stream: it is done with it once the handler has returned
	// and the response has been ended, so this is where the entry goes.
	defer m.conn.takeStream(streamId)
	defer inbound.Close()
	// Releases the stream's context once the handler is done with it, whether or
	// not the peer ever reset the stream.
	defer inbound.cancel()

	req, err := http.ReadRequest(bufio.NewReader(inbound))
	if err != nil {
		seedlog.Warnf("bidirequest: failed to read request on stream %d: %v", streamId, err)
		m.conn.sendReset(streamId)
		return
	}
	req.Body = &serverRequestBody{ReadCloser: req.Body, inbound: inbound}
	defer req.Body.Close()

	// The handler runs under the stream's context, so a peer that resets the
	// stream — because its caller cancelled, or the connection broke — cancels
	// the handler too.
	req = req.WithContext(inbound.ctx)

	// The message on the wire is HTTP/1.1, but the stream carrying it is full
	// duplex: the peer can keep sending body frames while this handler is
	// already sending response frames. connect refuses bidi streams on anything
	// below HTTP/2 because an HTTP/1.1 connection cannot do that — a check this
	// transport genuinely satisfies. Saying so here is what lets a bidi handler
	// run. The fields are never serialized; the WebSocket underneath is still an
	// HTTP/1.1 upgrade.
	req.Proto = "HTTP/2.0"
	req.ProtoMajor = 2
	req.ProtoMinor = 0
	req.RemoteAddr = "bidirequest"

	writer := &frameResponseWriter{
		conn:     m.conn,
		streamId: streamId,
		header:   http.Header{},
	}
	m.handler.ServeHTTP(writer, req)
	writer.finish()
}

// dispatch routes one received frame to the stream it belongs to, opening a
// stream — and the goroutine that serves it — when the peer initiates one.
//
// A stream outlives the end of either direction: the peer may have finished
// sending its request long before it resets the stream because its caller gave
// up. The entry is removed by whoever owns the stream — serve, once its handler
// has returned, or RoundTrip, once its response body is done — never here on a
// mere end-of-direction.
func (m *MuxTransport) dispatch(payload *bidirequestpb.Payload) {
	streamId := payload.GetStreamId()

	inbound := m.conn.getStream(streamId)
	if inbound == nil {
		if streamId&1 == m.side {
			// A frame for a stream this side opened and has since abandoned.
			// The peer will see the reset we sent; drop what is still in flight.
			return
		}
		if len(payload.GetData()) == 0 {
			// The bare tail of a stream already served and torn down. Opening
			// one here would run a handler on a request that never arrives.
			// A frame carrying data opens a stream even when it also ends it —
			// that is exactly what a request with no body looks like.
			return
		}
		inbound = newInboundStream()
		m.conn.putStream(streamId, inbound)
		go m.serve(streamId, inbound)
	}

	if payload.GetReset_() {
		m.conn.takeStream(streamId)
		inbound.finish(errStreamReset)
		return
	}
	if len(payload.GetData()) > 0 {
		inbound.push(payload.GetData())
	}
	if payload.GetEnd() {
		inbound.finish(nil)
	}
}

// startReceiveLoop drains the wire for the life of the connection, dispatching
// each frame to its stream. A receive error is terminal: it fails every stream
// still open, and no more frames will arrive.
func (m *MuxTransport) startReceiveLoop() {
	go func() {
		for {
			payload, err := m.conn.stream.Receive()
			if err != nil {
				for _, inbound := range m.conn.takeAllStreams(err) {
					inbound.finish(err)
				}
				return
			}
			m.dispatch(payload)
		}
	}()
}

// RoundTrip opens a stream to the peer and runs req on it. It returns as soon
// as the peer's response head has arrived, while the request body may still be
// going out and the response body still coming in — a full-duplex caller reads
// and writes against the two at once.
//
// The returned response owns the stream: closing its body is what releases the
// stream, and what tells the peer to stop sending if it has not finished.
func (m *MuxTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	streamId := m.nextId.Add(2) - 2

	inbound, err := m.conn.openInbound(streamId)
	if err != nil {
		return nil, err
	}

	// The request is written on its own goroutine: a full-duplex caller expects
	// the response to come back while it is still sending, so RoundTrip must not
	// wait for the body to finish before it returns.
	go func() {
		writeErr := m.writeRequest(streamId, req)
		if writeErr != nil {
			seedlog.Warnf("bidirequest: failed to write request on stream %d: %v", streamId, writeErr)
		}
	}()

	type readResult struct {
		resp *http.Response
		err  error
	}
	responded := make(chan readResult, 1)
	go func() {
		resp, readErr := http.ReadResponse(bufio.NewReader(inbound), req)
		responded <- readResult{resp: resp, err: readErr}
	}()

	select {
	case <-req.Context().Done():
		m.conn.takeStream(streamId)
		inbound.Close()
		m.conn.sendReset(streamId)
		return nil, req.Context().Err()
	case result := <-responded:
		if result.err != nil {
			m.conn.takeStream(streamId)
			inbound.Close()
			return nil, result.err
		}
		// See the note in serve: the response travelled a full-duplex stream,
		// and connect checks the version on this side too.
		result.resp.Proto = "HTTP/2.0"
		result.resp.ProtoMajor = 2
		result.resp.ProtoMinor = 0
		result.resp.Body = &responseBody{
			ReadCloser: result.resp.Body,
			conn:       m.conn,
			streamId:   streamId,
			inbound:    inbound,
		}
		return result.resp, nil
	}
}

var _ http.RoundTripper = (*MuxTransport)(nil)

// wrapClient builds the transport for one end of a connection and starts
// serving the peer's streams against handler. side seeds the stream ids this
// end will open, so that the two ends never pick the same one: 2, 4, 6, ... on
// the server side and 1, 3, 5, ... on the client side.
func wrapClient(stream PayloadStream, handler http.Handler, side uint32) *http.Client {
	transport := &MuxTransport{
		conn: &MuxConnection{
			stream:  stream,
			streams: map[uint32]*inboundStream{},
		},
		handler: handler,
		side:    side,
	}
	initialId := uint32(side)
	if initialId == 0 {
		initialId += 2
	}
	transport.nextId.Store(initialId)
	transport.startReceiveLoop()
	return &http.Client{Transport: transport}
}

// WrapClientSide turns the dialing end of a connection into an http.Client that
// reaches the peer, and serves the requests the peer sends back against
// handler. It returns immediately; the connection is served in the background
// until stream fails or is closed.
func WrapClientSide(stream PayloadStream, handler http.Handler) *http.Client {
	return wrapClient(stream, handler, CLIENT_SIDE)
}

// WrapServerSide is WrapClientSide for the accepting end of a connection: the
// http.Client it returns is how a server reaches back into the client that
// connected to it, which is the whole point of the tunnel.
func WrapServerSide(stream PayloadStream, handler http.Handler) *http.Client {
	return wrapClient(stream, handler, SERVER_SIDE)
}
