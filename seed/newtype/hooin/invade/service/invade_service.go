package service

import (
	"context"
	"errors"
	"io"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/keycloaklogin"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/invade/proto/invadepb"
	"github.com/ndscm/theseed/seed/newtype/hooin/onsite"
	"google.golang.org/grpc/codes"
	"google.golang.org/protobuf/types/known/emptypb"
)

// HooinInvadeService reaches into a person's workstation from outside it.
//
// The terminals it opens are held by the duty of the person whose workstation
// runs them, not here: a terminal is a shell on that workstation, and it dies
// with the commute connection that carries it.
type HooinInvadeService struct {
	office *onsite.Office
}

// StartTerminal opens a terminal and streams its output for as long as it
// lives. Keystrokes travel the other way as SendTerminalInput calls naming the
// same person and session, so the session outlives those requests but not this
// one: when this stream ends, the terminal is gone.
func (svc *HooinInvadeService) StartTerminal(
	ctx context.Context,
	req *connect.Request[invadepb.StartTerminalRequest],
	stream *connect.ServerStream[terminalpb.TerminalOutputFrame],
) error {
	loginUser, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}

	if req.Msg.GetPersonId() == "" {
		return seederr.CodeErrorf(codes.InvalidArgument, "person_id is required")
	}
	// Who may invade whom is a keycloak role for now: one role per person that
	// may be invaded, named after them. The empty client id asks for the role
	// under the client the token was issued to, which is the one that sent the
	// caller here.
	err = keycloaklogin.VerifyRole(loginUser, "", "invade:"+req.Msg.GetPersonId())
	if err != nil {
		return seederr.Wrap(err)
	}

	if req.Msg.GetStart().GetStart() == nil {
		return seederr.CodeErrorf(codes.InvalidArgument, "start is required")
	}
	// The agent ignores the session uuid — a stream carries one terminal, so it
	// has nothing to tell apart. Hooin does not: it holds a terminal per stream
	// but many streams at once, and SendTerminalInput finds one by this name
	// alone. So here it has to be there.
	if req.Msg.GetStart().GetSessionUuid() == "" {
		return seederr.CodeErrorf(codes.InvalidArgument, "start.session_uuid is required")
	}

	// TODO(nagi): add fine-grained authorization
	duty := svc.office.GetDuty(req.Msg.GetPersonId())
	if duty == nil {
		return seederr.CodeErrorf(codes.FailedPrecondition,
			"person %q is not on duty", req.Msg.GetPersonId())
	}

	session, err := duty.StartTerminal(ctx, loginUser.Sub, req.Msg.GetStart())
	if err != nil {
		return seederr.Wrap(err)
	}
	defer session.Close()

	for {
		frame, err := session.Receive()
		if errors.Is(err, io.EOF) {
			return nil
		}
		if err != nil {
			return seederr.Wrap(err)
		}

		// The front-end reads the very frames the agent sends: terminal output
		// is the same on both sides of hooin, so there is nothing to translate.
		err = stream.Send(frame)
		if err != nil {
			return seederr.Wrap(err)
		}
		if frame.GetExit() != nil {
			return nil
		}
	}
}

// SendTerminalInput types at an open terminal.
func (svc *HooinInvadeService) SendTerminalInput(
	ctx context.Context,
	req *connect.Request[invadepb.SendTerminalInputRequest],
) (*connect.Response[emptypb.Empty], error) {
	loginUser, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	if req.Msg.GetPersonId() == "" {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "person_id is required")
	}
	// The role is not checked again here, on every keystroke. StartTerminal
	// checked it, and a caller who failed that check has no terminal to type at:
	// the GetTerminal below is what stands in for the check, and it is a lookup
	// this call has to do anyway.

	sessionUuid := req.Msg.GetFrame().GetSessionUuid()
	if sessionUuid == "" {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "frame.session_uuid is required")
	}
	if req.Msg.GetFrame().GetStart() != nil {
		return nil, seederr.CodeErrorf(codes.InvalidArgument,
			"only StartTerminal may open a terminal")
	}

	duty := svc.office.GetDuty(req.Msg.GetPersonId())
	if duty == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition,
			"person %q is not on duty", req.Msg.GetPersonId())
	}

	// Knowing a session uuid is not enough. It is unguessable, but a terminal is
	// a shell: the person who opened it is the only one who may type at it. That
	// is why the terminal is asked for by owner — another subject's session id
	// finds nothing here, rather than being found and then refused.
	//
	// Nothing orders this call against the StartTerminal that opens the terminal,
	// so input that overtakes it finds nothing and is lost. That is what the
	// frame index is for: input is delivered, never queued, and a caller that
	// loses a keystroke has lost a keystroke rather than the shell.
	session, err := duty.GetTerminal(loginUser.Sub, sessionUuid)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// The frame reaches the agent as the caller wrote it. Its fields are not a
	// oneof, so it may carry both a resize and the keystrokes typed after it,
	// and the agent applies them in that order; forwarding it whole is what
	// keeps them together.
	err = session.Send(req.Msg.GetFrame())
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return connect.NewResponse(&emptypb.Empty{}), nil
}

// NewHooinInvadeService reaches workstations through the given office: it is
// where the duty of a person on it is found, and there is nothing to invade
// through but a duty.
func NewHooinInvadeService(office *onsite.Office) *HooinInvadeService {
	return &HooinInvadeService{
		office: office,
	}
}
