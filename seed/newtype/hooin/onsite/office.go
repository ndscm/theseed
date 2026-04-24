package onsite

import (
	"sync"

	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team"
	"google.golang.org/grpc/codes"
)

type Office struct {
	Team team.Team

	dutiesMutex sync.Mutex
	duties      map[string]*PersonDuty

	// stepSubscribers is keyed by subscriber pointer so that an
	// unsubscribing caller can remove itself in O(1) via `delete(...)`
	// without scanning the collection. Fanout iterates all entries
	// regardless, so the loss of ordering and the slightly higher
	// per-iteration cost compared to a slice is an acceptable tradeoff
	// for cheap unsubscribe.
	stepSubscribersMutex sync.Mutex
	stepSubscribers      map[*StepSubscriber]struct{}
}

func NewOffice(t team.Team) (*Office, error) {
	ofc := &Office{}
	ofc.Team = t
	ofc.duties = map[string]*PersonDuty{}
	ofc.stepSubscribers = map[*StepSubscriber]struct{}{}
	return ofc, nil
}

func (ofc *Office) GetDuty(person string) *PersonDuty {
	ofc.dutiesMutex.Lock()
	defer ofc.dutiesMutex.Unlock()
	return ofc.duties[person]
}

// SetDuty atomically claims the duty slot for person. It fails if the
// person is already on duty, so a stale commute goroutine's cleanup
// cannot delete a newer session's duty (which would silently orphan the
// new agent).
func (ofc *Office) SetDuty(person string, duty *PersonDuty) error {
	ofc.dutiesMutex.Lock()
	defer ofc.dutiesMutex.Unlock()
	_, exist := ofc.duties[person]
	if exist {
		return seederr.CodeErrorf(codes.AlreadyExists, "person %q is already on duty", person)
	}
	ofc.duties[person] = duty
	return nil
}

func (ofc *Office) ClearDuty(person string) {
	ofc.dutiesMutex.Lock()
	defer ofc.dutiesMutex.Unlock()
	delete(ofc.duties, person)
}

func (ofc *Office) SubscribeSteps(sub *StepSubscriber) {
	ofc.stepSubscribersMutex.Lock()
	defer ofc.stepSubscribersMutex.Unlock()
	ofc.stepSubscribers[sub] = struct{}{}
}

func (ofc *Office) UnsubscribeSteps(sub *StepSubscriber) {
	ofc.stepSubscribersMutex.Lock()
	defer ofc.stepSubscribersMutex.Unlock()
	delete(ofc.stepSubscribers, sub)
}

// matchStepSubscribers returns a snapshot of step subscribers matching
// the given person and topic. An empty subscriber personId matches any
// person; an empty subscriber topic matches any topic.
//
// Callers must not hold stepSubscribersMutex when sending on the
// returned subscribers' channels: the snapshot exists precisely so that
// the lock can be released before any blocking send.
func (ofc *Office) matchStepSubscribers(personId string, topic string) []*StepSubscriber {
	ofc.stepSubscribersMutex.Lock()
	defer ofc.stepSubscribersMutex.Unlock()
	targets := make([]*StepSubscriber, 0, len(ofc.stepSubscribers))
	for sub := range ofc.stepSubscribers {
		if sub.personId != "" && sub.personId != personId {
			continue
		}
		if sub.topic != "" && sub.topic != topic {
			continue
		}
		targets = append(targets, sub)
	}
	return targets
}

func (ofc *Office) BroadcastStep(personId string, topic string, step *brainpb.BrainStep) {
	subscribers := ofc.matchStepSubscribers(personId, topic)

	// Fanout must not block the reporting RPC on any one slow or
	// disappearing subscriber, so the send is non-blocking: if the
	// subscriber's channel can't accept the step right now, drop and
	// log. A cancelled subscriber will also land in this branch (its
	// channel stops being drained) and be cleaned up when its RPC
	// returns and calls UnsubscribeSteps.
	for _, sub := range subscribers {
		select {
		case sub.channel <- step:
		default:
			seedlog.Warnf("BrainStep subscriber channel full or gone, dropping step: person=%q topic=%q", personId, topic)
		}
	}
}

// DispatchBrainInput sends brainInput on personId's commute stream, under
// that duty's stream mutex. It returns FailedPrecondition if the person has
// no active commute session.
func (ofc *Office) DispatchBrainInput(personId string, brainInput *brainpb.BrainInput) error {
	if brainInput.GetUuid() == "" {
		brainInput.Uuid = uuid.NewString()
	}

	duty := ofc.GetDuty(personId)
	if duty == nil {
		return seederr.CodeErrorf(codes.FailedPrecondition, "person %q is not on duty", personId)
	}

	err := duty.Send(brainInput)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}
