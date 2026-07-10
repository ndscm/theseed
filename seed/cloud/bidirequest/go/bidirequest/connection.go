package bidirequest

import (
	"context"
	"errors"
	"io"
	"strconv"
	"sync"
	"sync/atomic"

	"github.com/ndscm/theseed/seed/cloud/bidirequest/proto/bidirequestpb"
)

// bodyChunkSize bounds how much of a message body one frame carries. A reader
// that offers less — a terminal delivering a keystroke, a connect stream
// delivering one envelope — is framed as it comes, so a small write crosses the
// connection immediately rather than waiting for a buffer to fill.
const bodyChunkSize = 32 * 1024

// inboundBacklog is how many frames may pile up for a stream whose reader is
// not keeping up. Beyond it the receive loop blocks, which stalls every other
// stream on the connection: there is one WebSocket and one loop draining it.
// The backlog exists to absorb bursts, not to make head-of-line blocking
// impossible — the peers here exchange small messages, and a reader that stops
// reading for long enough to fill it has stopped working.
const inboundBacklog = 64

var errStreamReset = errors.New("bidirequest: stream reset by peer")
var errStreamClosed = errors.New("bidirequest: stream closed")

// inboundStream is the reading end of one direction of a stream: the bytes the
// peer sends on it, handed to whoever is parsing the message.
//
// It is a channel rather than an io.Pipe because the receive loop must not
// block on a reader that has gone away, and because a pipe would force the loop
// to wait for every byte to be consumed before reading the next frame.
type inboundStream struct {
	chunks chan []byte

	// ctx is cancelled when the stream fails — the peer reset it, or the
	// connection broke. A handler serving the stream runs under it, so a caller
	// that gives up releases the handler rather than leaving it parked forever
	// on a request nobody is waiting for.
	//
	// A clean end does not cancel it: the request body is finished, but the
	// handler is still entitled to write its response.
	ctx    context.Context
	cancel context.CancelFunc

	// closed is shut by the consumer when it abandons the stream, releasing the
	// receive loop if it is parked handing over a frame nobody will read.
	closed    chan struct{}
	closeOnce sync.Once

	// ended records that the producer said its piece, so a consumer closing
	// afterwards knows not to reset a stream that already finished.
	ended atomic.Bool

	// finished guards the one-shot close of chunks. A peer that ends a stream
	// and then resets it — or that keeps sending after its own end — must not
	// close the channel twice or send on a closed one.
	finished     atomic.Bool
	finishedOnce sync.Once

	errMutex sync.Mutex
	err      error

	// cur is the remainder of the frame currently being handed out.
	cur []byte
}

func (s *inboundStream) setErr(err error) {
	s.errMutex.Lock()
	defer s.errMutex.Unlock()
	if s.err == nil {
		s.err = err
	}
}

func (s *inboundStream) readErr() error {
	s.errMutex.Lock()
	defer s.errMutex.Unlock()
	if s.err == nil {
		return io.EOF
	}
	return s.err
}

func (s *inboundStream) Read(p []byte) (int, error) {
	for len(s.cur) == 0 {
		select {
		case chunk, ok := <-s.chunks:
			if !ok {
				return 0, s.readErr()
			}
			s.cur = chunk
		case <-s.closed:
			return 0, errStreamClosed
		}
	}
	n := copy(p, s.cur)
	s.cur = s.cur[n:]
	return n, nil
}

// Close abandons the stream from the reader's side.
func (s *inboundStream) Close() error {
	s.closeOnce.Do(func() {
		close(s.closed)
	})
	return nil
}

// push hands a frame to the reader. It is only ever called from the receive
// loop, which is also the only caller of finish — so the channel is closed by
// exactly one goroutine and a send can never race that close.
func (s *inboundStream) push(data []byte) {
	if s.finished.Load() {
		return
	}
	select {
	case s.chunks <- data:
	case <-s.closed:
	}
}

// finish reports that no more frames will arrive. err is nil for a clean end,
// in which case the reader sees io.EOF once it has drained what was queued;
// anything else fails the stream and cancels whatever is serving it.
func (s *inboundStream) finish(err error) {
	if err == nil {
		s.ended.Store(true)
	} else {
		s.cancel()
	}
	s.setErr(err)
	s.finishedOnce.Do(func() {
		s.finished.Store(true)
		close(s.chunks)
	})
}

func newInboundStream() *inboundStream {
	ctx, cancel := context.WithCancel(context.Background())
	return &inboundStream{
		chunks: make(chan []byte, inboundBacklog),
		ctx:    ctx,
		cancel: cancel,
		closed: make(chan struct{}),
	}
}

// MuxConnection is the shared lower half of the transport: the one PayloadStream on
// the wire and the table of streams multiplexed over it. It owns sending frames
// and tracking inbound streams, so the per-stream types that only need to send
// or to release a stream — frameResponseWriter, responseBody — can depend on it
// without reaching for the whole MuxTransport.
type MuxConnection struct {
	stream PayloadStream

	sendMutex sync.Mutex

	streamsMutex sync.Mutex
	streams      map[uint32]*inboundStream

	recvErr error
}

func (c *MuxConnection) send(payload *bidirequestpb.Payload) error {
	c.sendMutex.Lock()
	defer c.sendMutex.Unlock()
	return c.stream.Send(payload)
}

// sendChunk frames data as one HTTP/1.1 chunk. Building the chunk header, the
// data, and its terminator into a single frame keeps a message body's framing
// intact no matter how the frames are interleaved with other streams'.
func (c *MuxConnection) sendChunk(streamId uint32, data []byte) error {
	buf := make([]byte, 0, len(data)+16)
	buf = strconv.AppendUint(buf, uint64(len(data)), 16)
	buf = append(buf, '\r', '\n')
	buf = append(buf, data...)
	buf = append(buf, '\r', '\n')
	return c.send(&bidirequestpb.Payload{StreamId: streamId, Data: buf})
}

// sendLastChunk writes the zero-length chunk that terminates a chunked body,
// and ends the sender's direction of the stream.
func (c *MuxConnection) sendLastChunk(streamId uint32) error {
	return c.send(&bidirequestpb.Payload{
		StreamId: streamId,
		Data:     []byte("0\r\n\r\n"),
		End:      true,
	})
}

func (c *MuxConnection) sendReset(streamId uint32) error {
	return c.send(&bidirequestpb.Payload{StreamId: streamId, Reset_: true})
}

func (c *MuxConnection) putStream(streamId uint32, inbound *inboundStream) {
	c.streamsMutex.Lock()
	defer c.streamsMutex.Unlock()
	c.streams[streamId] = inbound
}

func (c *MuxConnection) getStream(streamId uint32) *inboundStream {
	c.streamsMutex.Lock()
	defer c.streamsMutex.Unlock()
	return c.streams[streamId]
}

func (c *MuxConnection) takeStream(streamId uint32) *inboundStream {
	c.streamsMutex.Lock()
	defer c.streamsMutex.Unlock()
	inbound, exist := c.streams[streamId]
	if !exist {
		return nil
	}
	delete(c.streams, streamId)
	return inbound
}

func (c *MuxConnection) takeAllStreams(recvErr error) []*inboundStream {
	c.streamsMutex.Lock()
	defer c.streamsMutex.Unlock()
	c.recvErr = recvErr
	inbounds := make([]*inboundStream, 0, len(c.streams))
	for _, inbound := range c.streams {
		inbounds = append(inbounds, inbound)
	}
	c.streams = map[uint32]*inboundStream{}
	return inbounds
}

// openInbound registers the reading end of a stream this side is initiating.
func (c *MuxConnection) openInbound(streamId uint32) (*inboundStream, error) {
	c.streamsMutex.Lock()
	defer c.streamsMutex.Unlock()
	if c.recvErr != nil {
		return nil, c.recvErr
	}
	inbound := newInboundStream()
	c.streams[streamId] = inbound
	return inbound, nil
}
