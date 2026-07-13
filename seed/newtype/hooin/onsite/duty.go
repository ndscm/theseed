package onsite

import (
	"context"
	"net/http"
	"sync"

	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidirequest"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/terminal/proto/terminalpb"
	"github.com/ndscm/theseed/seed/newtype/amadeus/commute/client/go/commuteclient"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

// PersonDuty pairs a commute stream with a stream mutex.
//
// connect.ServerStream.Send is not safe for concurrent use, so every
// goroutine forwarding a BrainInput to the agent must hold streamMutex.
type PersonDuty struct {
	mutex  sync.Mutex
	stream bidirequest.PayloadStream
	client *http.Client

	// terminals are the terminals open on this person's workstation. They
	// belong to the duty because the workstation does: the person goes off
	// duty, the commute connection goes with them, and every shell it was
	// holding open is gone. Terminals are named by session uuid within one
	// duty, which is all a caller who has named a person needs to say.
	terminals *TerminalRegistry
}

func (d *PersonDuty) Send(ctx context.Context, brainInput *brainpb.BrainInput) error {
	d.mutex.Lock()
	defer d.mutex.Unlock()
	client := commuteclient.NewAmadeusCommuteClient(commuteclient.WithHttpClient(d.client))
	err := client.SendBrainInput(ctx, brainInput)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// StartTerminal opens a terminal on the person's workstation, owned by
// ownerSub, and holds it under the session uuid that start names — which is
// what the input typed at it later arrives by. The terminal lives until Close.
//
// The opening frame travels to the agent as the caller wrote it: everything on
// it — the session it names, the size it asks for — is the agent's.
//
// Unlike Send, it holds no lock: the commute connection multiplexes concurrent
// requests, so callers may have several in flight at once. A terminal needs
// exactly that — it holds one stream open for as long as its shell lives, and
// serializing behind it would stall every BrainInput for the same person.
func (d *PersonDuty) StartTerminal(
	ctx context.Context, owner string, start *terminalpb.TerminalInputFrame,
) (*TerminalSession, error) {
	client := commuteclient.NewAmadeusCommuteClient(commuteclient.WithHttpClient(d.client))
	stream := client.StartTerminal(ctx)

	// The start frame goes first, before the session is registered: until the
	// agent has been told to open a terminal, there is nothing for a keystroke
	// that found the session to type at.
	err := stream.Send(start)
	if err != nil {
		stream.CloseRequest()
		stream.CloseResponse()
		return nil, seederr.Wrap(err)
	}

	session, err := d.terminals.start(owner, start.GetSessionUuid(), stream)
	if err != nil {
		// The session id was taken, so the terminal under it is somebody else's:
		// close the stream this call opened, and leave the registry as it was
		// found.
		stream.CloseRequest()
		stream.CloseResponse()
		return nil, seederr.Wrap(err)
	}
	return session, nil
}

// GetTerminal returns the terminal on this workstation that owner opened under
// sessionUuid. A terminal somebody else opened is not owner's to find, and one
// that is not open yet is not open: nothing here waits for it to be.
func (d *PersonDuty) GetTerminal(owner string, sessionUuid string) (*TerminalSession, error) {
	session, err := d.terminals.get(owner, sessionUuid)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return session, nil
}

// SimpleFileSystem reaches the file system of this person's workstation.
//
// Like StartTerminal, and unlike Send, it holds no lock: the commute connection
// multiplexes concurrent requests, and an editor asks about many paths at once.
// Nothing here is held between two calls — a file is named, read, and let go —
// so a client per call costs nothing and keeps none of them waiting on another.
//
// What the file system allows is the workstation's own business: the agent runs
// each of these inside the playpen container as the person whose workstation it
// is. This forwards, and carries back what the file system said.
func (d *PersonDuty) SimpleFileSystem() *commuteclient.AmadeusCommuteClient {
	return commuteclient.NewAmadeusCommuteClient(commuteclient.WithHttpClient(d.client))
}

func CreatePersonDuty(stream bidirequest.PayloadStream, handler http.Handler) *PersonDuty {
	client := bidirequest.WrapServerSide(stream, handler)
	return &PersonDuty{
		stream:    stream,
		client:    client,
		terminals: newTerminalRegistry(),
	}
}
