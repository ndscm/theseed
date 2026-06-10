package onsite

import (
	"context"
	"net/http"
	"sync"

	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidirequest"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
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

func CreatePersonDuty(stream bidirequest.PayloadStream, handler http.Handler) *PersonDuty {
	client := bidirequest.WrapServerSide(stream, handler)
	return &PersonDuty{
		stream: stream,
		client: client,
	}
}
