package onsite

import (
	"sync"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

// PersonDuty pairs a commute stream with a stream mutex.
//
// connect.ServerStream.Send is not safe for concurrent use, so every
// goroutine forwarding a BrainInput to the agent must hold streamMutex.
type PersonDuty struct {
	streamMutex sync.Mutex
	stream      *connect.ServerStream[brainpb.BrainInput]
}

func NewPersonDuty(stream *connect.ServerStream[brainpb.BrainInput]) *PersonDuty {
	return &PersonDuty{
		stream: stream,
	}
}

func (d *PersonDuty) Send(brainInput *brainpb.BrainInput) error {
	d.streamMutex.Lock()
	defer d.streamMutex.Unlock()
	return d.stream.Send(brainInput)
}

// Flush forces the underlying HTTP response headers to be written to
// the client without sending a BrainInput. connect-go treats Send(nil)
// as a no-op that only flushes headers, which is how the hooin service
// lets a Commute client know its session is established.
func (d *PersonDuty) Flush() error {
	d.streamMutex.Lock()
	defer d.streamMutex.Unlock()
	return d.stream.Send(nil)
}
