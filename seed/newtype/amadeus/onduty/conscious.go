package onduty

import (
	"context"
	"errors"
	"net"
	"net/http"
	"sync"
	"time"

	"github.com/ndscm/theseed/seed/cloud/bidirequest/client/go/bidirequestclient"
	"github.com/ndscm/theseed/seed/cloud/bidirequest/go/bidirequest"
	"github.com/ndscm/theseed/seed/cloud/login/go/siliconlogin"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/amadeus/brain"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/client/go/commuteclient"
	"google.golang.org/grpc/codes"
)

// A broken connect stream is retried with exponential backoff, capped, until
// it comes back or Hibernate cancels the connect context.
const InitialBackoff = 1 * time.Second
const MaxBackoff = 10 * time.Minute
const MinStableConnection = 1 * time.Minute

func increaseBackoff(backoff time.Duration) time.Duration {
	if backoff == 0 {
		return InitialBackoff
	}
	backoff *= 2
	if backoff > MaxBackoff {
		return MaxBackoff
	}
	return backoff
}

// Conscious owns a single bidirectional connection to the commute server for
// the duration of a Wake → Hibernate cycle. It feeds brain input over that
// connection and reports brain steps back, transparently reconnecting if the
// stream breaks so registered topics survive the gap.
type Conscious struct {
	brain brain.Brain

	connectHandler http.Handler

	connectMutex      sync.Mutex
	connectCtx        context.Context
	cancelConnect     context.CancelFunc
	connectStream     bidirequest.PayloadStream
	connectStreamDone <-chan struct{}
	connectDone       chan struct{}

	commuteClient *commuteclient.HooinCommuteClient

	topicsMutex sync.Mutex
	topics      map[string]*LiveTopic
}

func (s *Conscious) SetConnectHandler(handler http.Handler) {
	s.connectHandler = handler
}

// getCommuteClient returns the client wired to the live connect stream.
// Reporters read it on every step so they follow reconnects: a reconnect
// swaps in a fresh client and the next step picks it up without re-registering.
func (s *Conscious) getCommuteClient() *commuteclient.HooinCommuteClient {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()
	return s.commuteClient
}

func (s *Conscious) checkTopicStarted(topic string) bool {
	s.topicsMutex.Lock()
	defer s.topicsMutex.Unlock()
	_, ok := s.topics[topic]
	return ok
}

// ensureTopicStarted must be called from a single goroutine per Conscious instance.
// The check-then-act on s.topics is not atomic, so concurrent callers for the
// same new topic would both reach Brain.RegisterStepHandler and the second would
// fail with "step handler already registered".
func (s *Conscious) ensureTopicStarted(
	topic string,
) error {
	if s.checkTopicStarted(topic) {
		return nil
	}

	reporter := &StepReporter{conscious: s}
	err := s.brain.RegisterStepHandler(topic, reporter)
	if err != nil {
		return seederr.Wrap(err)
	}

	s.topicsMutex.Lock()
	defer s.topicsMutex.Unlock()
	s.topics[topic] = &LiveTopic{reporter: reporter}
	return nil
}

func (s *Conscious) Input(
	input *brainpb.BrainInput,
) error {
	seedlog.Infof("Processing conscious input. input=%v", input)

	topic := input.GetTopic()
	if topic == "" {
		topic = "default"
	}
	err := s.ensureTopicStarted(topic)
	if err != nil {
		return seederr.Wrap(err)
	}

	err = s.brain.Input(s.connectCtx, topic, input)
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

// connect opens a fresh connect stream and builds the commuteClient that rides it,
// storing both along with connectStreamDone, the channel that closes when the
// stream breaks. The caller must hold connectMutex. A fresh SiliconLogin runs
// per connect so reconnects pick up a refreshed access token rather than
// reusing the one baked in at Wake.
func (s *Conscious) connect() error {
	ctx, err := siliconlogin.SiliconLogin(s.connectCtx)
	if err != nil {
		return seederr.Wrap(err)
	}

	connectClient := bidirequestclient.NewBidirequestClient(commuteclient.HooinCommuteServiceServerFlag())
	connectStream, err := connectClient.Connect(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	muxClient := bidirequest.WrapClientSide(connectStream, s.connectHandler)

	s.commuteClient = commuteclient.NewHooinCommuteClient(
		commuteclient.WithHttpClient(muxClient),
	)
	s.connectStream = connectStream
	s.connectStreamDone = connectStream.Context().Done()

	return nil
}

// cleanupConnectStream closes the live stream, if any. Best effort: by the time
// it runs the peer has usually already gone away.
func (s *Conscious) cleanupConnectStream() {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()
	if s.connectStream != nil {
		// A broken stream's conn is usually already closed by the read loop, so
		// Close returns net.ErrClosed. That is the expected case here, not a
		// failure worth surfacing; only a different error is.
		err := s.connectStream.Close()
		if err != nil && !errors.Is(err, net.ErrClosed) {
			seedlog.Warnf("Failed to close connect stream. err=%v", err)
		}
	}
	s.commuteClient = nil
	s.connectStream = nil
	s.connectStreamDone = nil
}

// reconnect replaces a broken connect stream, retrying with backoff until it
// succeeds (storing the fresh stream and connectStreamDone) or the connect
// context is canceled. It returns nil once reconnected, or a non-nil error if
// the context was canceled meanwhile, telling commute to tear down instead of
// looping.
func (s *Conscious) reconnect() error {
	s.cleanupConnectStream()

	backoff := increaseBackoff(0)
	for {
		// Connect under connectMutex unless the connect context was already
		// canceled by Hibernate, in which case hibernated tells us to stop
		// retrying.
		hibernated, err := func() (bool, error) {
			s.connectMutex.Lock()
			defer s.connectMutex.Unlock()
			if s.connectCtx.Err() != nil {
				return true, nil
			}
			err := s.connect()
			if err != nil {
				return false, seederr.Wrap(err)
			}
			return false, nil
		}()
		if hibernated {
			return seederr.WrapErrorf("hibernated during reconnect")
		}
		if err != nil {
			seedlog.Warnf("Conscious reconnect failed, retrying in %v: %v", backoff, err)
			select {
			case <-s.connectCtx.Done():
				return seederr.WrapErrorf("hibernated during reconnect backoff")
			case <-time.After(backoff):
			}
			backoff = increaseBackoff(backoff)
			continue
		}
		break
	}
	seedlog.Infof("Conscious reconnected.")
	return nil
}

// shutdown closes the live stream, hibernates every topic, and clears the
// connection state, then signals connectDone so a waiting re-wake can proceed.
func (s *Conscious) shutdown() {
	s.cleanupConnectStream()

	func() {
		s.topicsMutex.Lock()
		defer s.topicsMutex.Unlock()
		wg := sync.WaitGroup{}
		for topic := range s.topics {
			wg.Add(1)
			go func(topic string) {
				defer wg.Done()
				err := s.brain.Hibernate(topic, true)
				if err != nil {
					seedlog.Errorf("Hibernate topic %q failed: %v", topic, err)
				}
			}(topic)
		}
		wg.Wait()
		s.topics = nil
	}()

	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()
	s.cancelConnect()
	close(s.connectDone)
	s.commuteClient = nil
	s.connectCtx = nil
	s.cancelConnect = nil
	s.connectStream = nil
	s.connectStreamDone = nil
	s.connectDone = nil
}

// commute supervises the connect stream for the life of a Wake. It tears down
// only when the connect context is canceled (an intentional Hibernate); a
// stream that breaks on its own is reconnected, keeping the live topics
// registered across the gap.
func (s *Conscious) commute() {
	defer s.shutdown()

	backoff := increaseBackoff(0)
	for {
		currentRound := time.Now()
		select {
		case <-s.connectCtx.Done():
			return
		case <-s.connectStreamDone:
			if time.Since(currentRound) >= MinStableConnection {
				backoff = increaseBackoff(0)
			} else {
				select {
				case <-s.connectCtx.Done():
					return
				case <-time.After(backoff):
				}
				backoff = increaseBackoff(backoff)
			}
			err := s.reconnect()
			if err != nil {
				return
			}
		}
	}
}

// Wake opens the connect stream and starts the commute goroutine that
// supervises it. It returns FailedPrecondition if already awake, or the dial
// error if the initial connection fails.
func (s *Conscious) Wake() error {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()

	// Wake is exclusive: a Wake cycle owns connectCtx from here until shutdown
	// clears it. connectStream is not a safe liveness sentinel — reconnect
	// transiently nils it mid-cycle — so guard on connectCtx, which stays set
	// across reconnects and is cleared only once the commute goroutine has fully
	// torn down. A Hibernate → Wake sequence can still observe connectCtx != nil
	// because Hibernate only cancels the context and returns; callers that need
	// a clean re-wake should wait on Hibernate's returned done channel first.
	if s.connectCtx != nil {
		return seederr.CodeErrorf(codes.FailedPrecondition, "already awake")
	}

	connectCtx, cancelConnect := context.WithCancel(context.Background())
	s.connectCtx = connectCtx
	s.cancelConnect = cancelConnect
	s.connectDone = make(chan struct{})
	s.topics = map[string]*LiveTopic{}

	err := s.connect()
	if err != nil {
		s.cancelConnect()
		s.connectCtx = nil
		s.cancelConnect = nil
		s.connectDone = nil
		s.topics = nil
		return seederr.Wrap(err)
	}

	go s.commute()

	return nil
}

// Hibernate cancels the connect context, signaling the commute goroutine to
// close the stream and hibernate every topic. It returns a channel that closes
// once that teardown completes, or nil if not awake. It does not block.
func (s *Conscious) Hibernate() chan struct{} {
	s.connectMutex.Lock()
	defer s.connectMutex.Unlock()
	if s.connectCtx == nil {
		return nil
	}
	s.cancelConnect()
	return s.connectDone
}

func CreateConscious() (*Conscious, error) {
	b, err := brain.DefaultBrain()
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return &Conscious{brain: b}, nil
}
