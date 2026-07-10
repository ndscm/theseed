package onsite

import (
	"sync"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"google.golang.org/grpc/codes"
)

// TerminalSession is one open terminal: the stream to the agent running it.
//
// Who may type at it is the registry's business rather than the session's — a
// terminal is found under the owner that opened it, so one that has been found
// at all is one the caller is allowed to type at.
type TerminalSession struct {
	// streamMutex serializes writes to the agent. Every keystroke arrives as its
	// own request, on its own goroutine, and connect streams are not safe for
	// concurrent Send.
	streamMutex sync.Mutex
	stream      *connect.BidiStreamForClient[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame]

	// onClosing runs as the terminal ends, before its stream is closed. It is
	// how whoever opened the terminal hears that it is over — the registry
	// holding it takes its name back this way — so the session itself has
	// nothing to know about who is keeping it.
	onClosing func()
}

// Send types at the terminal.
func (s *TerminalSession) Send(frame *terminalpb.TerminalInputFrame) error {
	s.streamMutex.Lock()
	defer s.streamMutex.Unlock()
	err := s.stream.Send(frame)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// Receive returns the next frame the terminal printed, and io.EOF once the
// agent has hung up.
func (s *TerminalSession) Receive() (*terminalpb.TerminalOutputFrame, error) {
	frame, err := s.stream.Receive()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return frame, nil
}

// Close ends the terminal: the shell dies with the stream that holds it open,
// and the session's name is free again.
func (s *TerminalSession) Close() {
	s.onClosing()
	s.stream.CloseRequest()
	s.stream.CloseResponse()
}

// TerminalRegistry holds one person's open terminals, by the owner that opened
// each and the session id it was opened under.
//
// Owner comes first because a terminal is a shell: the person who opened it is
// the only one who may type at it. Keeping each owner's terminals apart is what
// enforces that, rather than a check somebody has to remember to make — a
// session id looked up under one owner cannot find another owner's terminal,
// however the caller came by the id.
//
// Only a terminal that is open is here. A session id nobody has opened is
// nothing to hold on to, so naming one costs the registry nothing: it is
// answered on the spot, rather than kept and waited on.
type TerminalRegistry struct {
	mutex    sync.Mutex
	sessions map[string]map[string]*TerminalSession
}

// start holds stream as a terminal under owner's sessionUuid, and hands back
// the session that is now typed at through it. A session id the owner already
// has open is refused rather than stolen; two owners may hold the same id,
// since neither can reach the other's.
//
// The session is published with its stream already on it, so a keystroke that
// finds it has something to type at. Closing it takes its name back here, which
// is why the registry is what builds it rather than the caller: nobody else has
// to know how a terminal is let go of.
func (r *TerminalRegistry) start(
	owner string,
	sessionUuid string,
	stream *connect.BidiStreamForClient[terminalpb.TerminalInputFrame, terminalpb.TerminalOutputFrame],
) (*TerminalSession, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	ownerSessions, exist := r.sessions[owner]
	if !exist {
		ownerSessions = map[string]*TerminalSession{}
		r.sessions[owner] = ownerSessions
	}
	_, exist = ownerSessions[sessionUuid]
	if exist {
		return nil, seederr.CodeErrorf(codes.AlreadyExists, "terminal %q is already open", sessionUuid)
	}
	session := &TerminalSession{
		stream: stream,
		onClosing: func() {
			r.discard(owner, sessionUuid)
		},
	}
	ownerSessions[sessionUuid] = session
	return session, nil
}

// discard drops owner's sessionUuid, and the owner with it once they hold no
// terminals.
func (r *TerminalRegistry) discard(owner string, sessionUuid string) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	ownerSessions, exist := r.sessions[owner]
	if !exist {
		return
	}
	delete(ownerSessions, sessionUuid)
	if len(ownerSessions) == 0 {
		delete(r.sessions, owner)
	}
}

// get returns the terminal owner opened under sessionUuid. A terminal that is
// not open — including one another owner opened — is not found.
func (r *TerminalRegistry) get(owner string, sessionUuid string) (*TerminalSession, error) {
	r.mutex.Lock()
	defer r.mutex.Unlock()
	session, exist := r.sessions[owner][sessionUuid]
	if !exist {
		return nil, seederr.CodeErrorf(codes.NotFound, "terminal %q is not open", sessionUuid)
	}
	return session, nil
}

func newTerminalRegistry() *TerminalRegistry {
	return &TerminalRegistry{
		sessions: map[string]map[string]*TerminalSession{},
	}
}
