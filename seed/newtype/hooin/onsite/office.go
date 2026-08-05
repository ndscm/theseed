package onsite

import (
	"context"
	"sync"
	"time"

	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/gajetto/team"
	"github.com/ndscm/theseed/seed/newtype/steins/database/ent"
	"google.golang.org/grpc/codes"
)

type Office struct {
	Team team.Team

	db *ent.Client

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

// DB returns the ent client backing persisted BrainSteps. It may be nil
// when the office was created without a database.
func (ofc *Office) GetDatabase() *ent.Client {
	return ofc.db
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

// saveStep persists step to the database. It is intended to run in its
// own goroutine off the reporting RPC path, so it uses a fresh timeout
// ctx rather than the request ctx and only logs failures.
func (ofc *Office) saveStep(personId string, step *brainpb.BrainStep) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Minute)
	defer cancel()

	create := ofc.db.BrainStep.Create()
	stepUuid, err := uuid.Parse(step.GetUuid())
	if err != nil {
		seedlog.Warnf("Failed to parse BrainStep uuid %q: %v", step.GetUuid(), err)
	} else {
		create.SetID(stepUuid)
	}
	create.SetPersonID(personId)
	create.SetEmitTime(step.GetEmitTime().AsTime())
	create.SetType(step.GetType())
	create.SetTopic(step.GetTopic())
	create.SetThreadUUID(step.GetThreadUuid())
	create.SetData(step.GetData().AsMap())

	_, err = create.Save(ctx)
	if err != nil {
		seedlog.Warnf("Failed to save BrainStep to database: %v", err)
	}
}

// matchStepSubscribers returns a snapshot of step subscribers matching
// the given person and topic. An empty subscriber personId matches any
// person; an empty subscriber topic matches any topic.
//
// Callers must not hold stepSubscribersMutex when sending on the
// returned subscribers' channels: the snapshot exists precisely so that
// the lock can be released before any blocking send.
func (ofc *Office) matchStepSubscribers(personId string, topic string, threadUuid string) []*StepSubscriber {
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
		if sub.threadUuid != "" && sub.threadUuid != threadUuid {
			continue
		}
		targets = append(targets, sub)
	}
	return targets
}

func (ofc *Office) BroadcastStep(personId string, step *brainpb.BrainStep) {
	// Persist asynchronously: this runs on the reporting RPC path, and a
	// slow or stalled database must not block the RPC response nor delay
	// live-subscriber fanout below. saveStep uses its own timeout ctx,
	// independent of the request lifetime.
	if ofc.db != nil {
		go ofc.saveStep(personId, step)
	}
	subscribers := ofc.matchStepSubscribers(personId, step.GetTopic(), step.GetThreadUuid())

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
			seedlog.Warnf("BrainStep subscriber channel full or gone, dropping step: person=%q topic=%q", personId, step.GetTopic())
		}
	}
}

// DispatchBrainInput sends brainInput on personId's commute stream, under
// that duty's stream mutex. It returns FailedPrecondition if the person has
// no active commute session.
func (ofc *Office) DispatchBrainInput(ctx context.Context, personId string, brainInput *brainpb.BrainInput) error {
	if brainInput.GetUuid() == "" {
		brainInput.Uuid = uuid.NewString()
	}

	duty := ofc.GetDuty(personId)
	if duty == nil {
		return seederr.CodeErrorf(codes.FailedPrecondition, "person %q is not on duty", personId)
	}

	err := duty.Send(ctx, brainInput)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func CreateOffice(t team.Team, db *ent.Client) (*Office, error) {
	ofc := &Office{}
	ofc.Team = t
	ofc.duties = map[string]*PersonDuty{}
	ofc.stepSubscribers = map[*StepSubscriber]struct{}{}
	ofc.db = db
	return ofc, nil
}
