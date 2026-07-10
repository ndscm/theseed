package service

import (
	"context"
	"errors"
	"math"
	"sync"

	"connectrpc.com/connect"
	"github.com/creack/pty"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/onduty"
	"github.com/ndscm/theseed/seed/newtype/amadeus/playpen"
	"google.golang.org/grpc/codes"
)

// sanitizeWindowDimension narrows a window size carried as a proto uint32 to the
// uint16 a pseudo-terminal is sized in. A dimension that does not fit is a
// client bug, not something to silently truncate into a plausible-looking
// terminal.
func sanitizeWindowDimension(value uint32) (uint16, error) {
	if value > math.MaxUint16 {
		return 0, seederr.CodeErrorf(codes.InvalidArgument,
			"window dimension %d exceeds %d", value, math.MaxUint16)
	}
	return uint16(value), nil
}

// applyTerminalInputFrame delivers everything a frame carries to the terminal.
//
// The fields are not a oneof, so a frame may set several at once and each is a
// thing to do, not a case to choose between: a window that changed size and the
// keystrokes typed after it travel together. They are applied in the order the
// terminal would have seen them — the resize first, so the shell echoes those
// keystrokes at the size they were typed at.
//
// What a frame carries is told by presence. Bytes have none in proto3, which is
// why `input` is judged by length: an empty one is indistinguishable from an
// absent one, and is nothing to deliver either way. `start` is not applied
// here; only the frame that opens the terminal may carry one.
func applyTerminalInputFrame(terminal *playpen.PlaypenTerminal, frame *terminalpb.TerminalInputFrame) error {
	if resize := frame.GetResize(); resize != nil {
		rows, err := sanitizeWindowDimension(resize.GetRows())
		if err != nil {
			return seederr.Wrap(err)
		}
		cols, err := sanitizeWindowDimension(resize.GetCols())
		if err != nil {
			return seederr.Wrap(err)
		}
		err = terminal.Resize(rows, cols)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	if input := frame.GetInput(); len(input) > 0 {
		_, err := terminal.Stdin.Write(input)
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	return nil
}

// terminalOutputSender serializes and indexes the output frames the agent
// sends on a StartTerminal stream. Both the goroutine draining the pty and the
// one reading input write to it — connect streams are not safe for concurrent
// Send — and it stamps each frame with the next index so a driver can tell
// whether any were lost on the way.
type terminalOutputSender struct {
	stream *connect.BidiStream[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame]

	mutex     sync.Mutex
	nextIndex uint32
}

func (s *terminalOutputSender) send(frame *terminalpb.TerminalOutputFrame) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()
	frame.Index = s.nextIndex
	s.nextIndex++
	err := s.stream.Send(frame)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// terminalInputSorter tracks the next input index a session expects. Frames are
// not buffered to put them back in order: a gap means a frame was lost or
// overtaken in transit, and the expectation jumps past it — the missing frames
// are never waited for, and any that arrive after their turn are dropped.
type terminalInputSorter struct {
	next uint32
}

// validate advances the sorter past frame and reports whether to apply it. It
// returns an error whenever frame is out of order — behind what has already
// been applied, or ahead of the one expected — which the caller reports to the
// driver without ending the terminal.
//
// A frame behind is a duplicate, or a straggler whose turn already passed; its
// place in the input is gone, so it is dropped. A frame ahead means the ones
// between it and the one expected are lost — but the frame itself is real
// input, so it is applied and only the missing ones are given up. The sorter
// never waits for or reorders anything.
func (s *terminalInputSorter) validate(frame *terminalpb.TerminalInputFrame) (bool, error) {
	index := frame.GetIndex()
	switch {
	case index < s.next:
		return false, seederr.CodeErrorf(codes.DataLoss,
			"terminal input frame %d arrived after its turn, past %d", index, s.next)
	case index == s.next:
		s.next = index + 1
		return true, nil
	default:
		missing := index - s.next
		s.next = index + 1
		return true, seederr.CodeErrorf(codes.DataLoss,
			"terminal input frame %d arrived with %d frame(s) missing before it", index, missing)
	}
}

// toTerminalError renders err as a TerminalError to hand a driver. It carries
// the error's own code, defaulting to DataLoss, and its reason without the
// stack trace a SeedError's Error() would otherwise drag onto someone's screen.
func toTerminalError(err error) *terminalpb.TerminalError {
	terminalError := &terminalpb.TerminalError{
		Code:    int32(codes.DataLoss),
		Message: err.Error(),
	}
	seedErr := &seederr.SeedError{}
	if errors.As(err, &seedErr) {
		if seedErr.Code() != 0 {
			terminalError.Code = int32(seedErr.Code())
		}
		if cause := seedErr.Unwrap(); cause != nil {
			terminalError.Message = cause.Error()
		}
	}
	return terminalError
}

// openTerminal reads the opening frame off the stream and starts the terminal
// at the size it asks for, returning the frame so the caller can apply the rest
// of what it carries and order input from the one after it.
//
// The frame's session_uuid is ignored: a stream carries one terminal, so there
// is nothing to tell apart. Naming a session is for the driver's own
// bookkeeping — hooin holds many terminals and needs to know which of its
// callers a stream belongs to — and means nothing on this side.
func openTerminal(
	ctx context.Context,
	playpenController *playpen.PlaypenController,
	stream *connect.BidiStream[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame],
) (*playpen.PlaypenTerminal, *terminalpb.TerminalInputFrame, error) {
	frame, err := stream.Receive()
	if err != nil {
		return nil, nil, seederr.Wrap(err)
	}
	start := frame.GetStart()
	if start == nil {
		return nil, nil, seederr.CodeErrorf(codes.InvalidArgument, "first terminal frame must be a start")
	}
	rows, err := sanitizeWindowDimension(start.GetRows())
	if err != nil {
		return nil, nil, seederr.Wrap(err)
	}
	cols, err := sanitizeWindowDimension(start.GetCols())
	if err != nil {
		return nil, nil, seederr.Wrap(err)
	}

	window := pty.Winsize{Rows: rows, Cols: cols}
	terminal, err := playpenController.StartTerminal(ctx, window, "/usr/bin/zsh", []string{"-i"})
	if err != nil {
		return nil, nil, seederr.Wrap(err)
	}
	return terminal, frame, nil
}

// pumpTerminalInput forwards keystrokes and resizes to the terminal until the
// peer stops sending. The peer hanging up is how a terminal ends, so it closes
// the terminal on the way out, which releases the output pump.
//
// It applies frames in index order. A gap — an input frame lost or overtaken
// before it arrived — is reported to the driver as an `error` output frame, so
// it learns a keystroke went missing without the terminal ending over it.
func pumpTerminalInput(
	stream *connect.BidiStream[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame],
	terminal *playpen.PlaypenTerminal,
	sorter *terminalInputSorter,
	sender *terminalOutputSender,
) {
	defer terminal.Close()

	for {
		frame, err := stream.Receive()
		if err != nil {
			return
		}

		apply, err := sorter.validate(frame)
		if err != nil {
			sendErr := sender.send(&terminalpb.TerminalOutputFrame{
				Error: toTerminalError(err),
			})
			if sendErr != nil {
				seedlog.Errorf("Failed to report out-of-order terminal input: %v", sendErr)
				return
			}
		}
		if !apply {
			continue
		}

		if frame.GetStart() != nil {
			// Only the opening frame may start a terminal, but the rest of this
			// one still says something.
			seedlog.Warnf("Ignoring a second terminal start")
		}
		err = applyTerminalInputFrame(terminal, frame)
		if err != nil {
			seedlog.Errorf("Failed to apply terminal input frame: %v", err)
			return
		}
	}
}

func startTerminal(
	ctx context.Context,
	conscious *onduty.Conscious,
	stream *connect.BidiStream[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame],
) error {
	// shellChunkSize bounds one read off the terminal. A shell that writes less
	// than this — which is most of the time — is forwarded as it writes, so
	// output appears as it is produced rather than once a buffer fills.
	const shellChunkSize = 32 * 1024

	playpenController := conscious.GetPlaypenController()
	if playpenController == nil {
		return seederr.CodeErrorf(codes.FailedPrecondition, "no playpen container is running")
	}

	terminal, startFrame, err := openTerminal(ctx, playpenController, stream)
	if err != nil {
		return seederr.Wrap(err)
	}
	defer terminal.Close()

	// The opening frame may carry more than the start: `start` is a field, not
	// a oneof arm, and a driver that opens a terminal and types into it in one
	// breath is entitled to say so. It is applied before the input pump starts,
	// so it lands ahead of anything typed after it.
	err = applyTerminalInputFrame(terminal, startFrame)
	if err != nil {
		return seederr.Wrap(err)
	}

	sender := &terminalOutputSender{stream: stream}
	// The start frame is the baseline; input is ordered from the one after it.
	sorter := &terminalInputSorter{next: startFrame.GetIndex() + 1}
	go pumpTerminalInput(stream, terminal, sorter, sender)

	// The output pump owns the foreground. It ends when the terminal stops
	// producing — because the shell exited, or because the input pump hung the
	// terminal up after the peer went away.
	buf := make([]byte, shellChunkSize)
	for {
		n, readErr := terminal.Stdout.Read(buf)
		if n > 0 {
			sendErr := sender.send(&terminalpb.TerminalOutputFrame{Output: buf[:n]})
			if sendErr != nil {
				return seederr.Wrap(sendErr)
			}
		}
		if readErr != nil {
			break
		}
	}

	// Reap the shell for its exit status. A terminal hung up on purpose has
	// none worth reporting: the read error is the hangup the input pump caused.
	exitMessage := ""
	waitErr := terminal.Wait()
	if waitErr != nil && ctx.Err() == nil {
		exitMessage = waitErr.Error()
	}
	seedlog.Infof("Playpen terminal ended. err=%v", waitErr)

	err = sender.send(&terminalpb.TerminalOutputFrame{
		Exit: &terminalpb.TerminalError{Message: exitMessage},
	})
	if err != nil {
		return seederr.Wrap(err)
	}

	return nil
}
