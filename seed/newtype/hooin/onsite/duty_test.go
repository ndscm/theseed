package onsite

import (
	"context"
	"errors"
	"io"
	"net/http"
	"sync"
	"testing"
	"time"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidirequest"
	"github.com/ndscm/theseed/seed/cloud/bidirequest/proto/bidirequestpb"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/proto/commutepbconnect"
)

// pipeStream is one end of an in-memory PayloadStream pair, standing in for the
// WebSocket hooin and amadeus really share.
type pipeStream struct {
	in  <-chan *bidirequestpb.Payload
	out chan<- *bidirequestpb.Payload

	closeOnce sync.Once
	done      chan struct{}
}

func (s *pipeStream) Send(payload *bidirequestpb.Payload) error {
	select {
	case s.out <- payload:
		return nil
	case <-s.done:
		return io.ErrClosedPipe
	}
}

func (s *pipeStream) Receive() (*bidirequestpb.Payload, error) {
	select {
	case payload := <-s.in:
		return payload, nil
	case <-s.done:
		return nil, io.EOF
	}
}

func (s *pipeStream) Close() error {
	s.closeOnce.Do(func() {
		close(s.done)
	})
	return nil
}

func newPipeStreamPair() (*pipeStream, *pipeStream) {
	toHooin := make(chan *bidirequestpb.Payload, 256)
	toAmadeus := make(chan *bidirequestpb.Payload, 256)
	done := make(chan struct{})
	hooin := &pipeStream{in: toHooin, out: toAmadeus, done: done}
	amadeus := &pipeStream{in: toAmadeus, out: toHooin, done: done}
	return hooin, amadeus
}

// echoCommuteService stands in for the agent. Its StartTerminal answers each frame
// as it arrives, which nothing but a genuinely full-duplex stream can carry:
// the caller reads a reply while its own request is still open.
type echoCommuteService struct {
	commutepbconnect.UnimplementedAmadeusCommuteServiceHandler
}

func (s *echoCommuteService) StartTerminal(
	ctx context.Context,
	stream *connect.BidiStream[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame],
) error {
	first, err := stream.Receive()
	if err != nil {
		return err
	}
	if first.GetStart() == nil {
		return connect.NewError(connect.CodeInvalidArgument,
			errors.New("first terminal frame must be a start"))
	}

	for {
		frame, err := stream.Receive()
		if errors.Is(err, io.EOF) {
			break
		}
		if err != nil {
			return err
		}
		input := frame.GetInput()
		if input == nil {
			continue
		}
		err = stream.Send(&terminalpb.TerminalOutputFrame{
			Output: append([]byte("echo:"), input...),
		})
		if err != nil {
			return err
		}
	}

	return stream.Send(&terminalpb.TerminalOutputFrame{
		Exit: &terminalpb.TerminalError{},
	})
}

var _ commutepbconnect.AmadeusCommuteServiceHandler = (*echoCommuteService)(nil)

// commuteAgent wires an agent-side commute handler onto one end of a tunnel and
// returns a duty holding the hooin end — the production arrangement, with the
// WebSocket replaced by a channel pair.
func commuteAgent(t *testing.T) *PersonDuty {
	t.Helper()

	mux := http.NewServeMux()
	mux.Handle(commutepbconnect.NewAmadeusCommuteServiceHandler(&echoCommuteService{}))

	hooinStream, amadeusStream := newPipeStreamPair()
	t.Cleanup(func() {
		hooinStream.Close()
	})

	// The agent dials out, so it is the client side of the tunnel; hooin
	// answers, and reaches back down it.
	bidirequest.WrapClientSide(amadeusStream, mux)
	return CreatePersonDuty(hooinStream, http.NotFoundHandler())
}

// startFrame is the frame that opens a terminal under sessionUuid.
func startFrame(sessionUuid string) *terminalpb.TerminalInputFrame {
	return &terminalpb.TerminalInputFrame{
		SessionUuid: &sessionUuid,
		Start:       &terminalpb.TerminalWindowSize{Rows: 24, Cols: 80},
	}
}

func TestCommuteBidiStream(t *testing.T) {
	t.Run("replies arrive while the request is still open", func(t *testing.T) {
		duty := commuteAgent(t)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		session, err := duty.StartTerminal(ctx, "owner", startFrame("open-terminal"))
		if err != nil {
			t.Fatalf("StartTerminal: %v", err)
		}
		defer session.Close()

		err = session.Send(&terminalpb.TerminalInputFrame{Input: []byte("one")})
		if err != nil {
			t.Fatalf("Send first input: %v", err)
		}
		frame, err := session.Receive()
		if err != nil {
			t.Fatalf("Receive first output: %v", err)
		}
		if string(frame.GetOutput()) != "echo:one" {
			t.Fatalf("first output = %q", frame.GetOutput())
		}

		// The second exchange is only reachable because the request was never
		// closed: a half-duplex transport would have needed it drained first.
		err = session.Send(&terminalpb.TerminalInputFrame{Input: []byte("two")})
		if err != nil {
			t.Fatalf("Send second input: %v", err)
		}
		frame, err = session.Receive()
		if err != nil {
			t.Fatalf("Receive second output: %v", err)
		}
		if string(frame.GetOutput()) != "echo:two" {
			t.Errorf("second output = %q", frame.GetOutput())
		}
	})

	t.Run("half-closing the request ends the stream cleanly", func(t *testing.T) {
		duty := commuteAgent(t)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		session, err := duty.StartTerminal(ctx, "owner", startFrame("half-close"))
		if err != nil {
			t.Fatalf("StartTerminal: %v", err)
		}
		defer session.Close()

		// Half-closing is the transport's business rather than the session's,
		// which ends a terminal by closing both halves at once.
		err = session.stream.CloseRequest()
		if err != nil {
			t.Fatalf("CloseRequest: %v", err)
		}

		frame, err := session.Receive()
		if err != nil {
			t.Fatalf("Receive exit: %v", err)
		}
		if frame.GetExit() == nil {
			t.Fatalf("expected an exit frame, got %v", frame)
		}

		_, err = session.Receive()
		if !errors.Is(err, io.EOF) {
			t.Errorf("after exit, Receive err = %v, want io.EOF", err)
		}
	})

	t.Run("an error from the agent reaches the caller", func(t *testing.T) {
		duty := commuteAgent(t)

		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()

		// Opening with something other than a start is what the agent refuses.
		// The stream carries that refusal, so StartTerminal itself succeeds and
		// the caller hears about it on the first frame it waits for.
		sessionUuid := "premature"
		session, err := duty.StartTerminal(ctx, "owner", &terminalpb.TerminalInputFrame{
			SessionUuid: &sessionUuid,
			Input:       []byte("premature"),
		})
		if err != nil {
			t.Fatalf("StartTerminal: %v", err)
		}
		defer session.Close()

		_, err = session.Receive()
		if err == nil {
			t.Fatal("expected the agent to refuse the stream")
		}
		if connect.CodeOf(err) != connect.CodeInvalidArgument {
			t.Errorf("code = %v, want %v", connect.CodeOf(err), connect.CodeInvalidArgument)
		}
	})
}
