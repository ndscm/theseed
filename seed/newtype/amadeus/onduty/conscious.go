package onduty

import (
	"context"
	"errors"
	"sync"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/client/go/commuteclient"
	"google.golang.org/grpc/codes"
)

type Conscious struct {
	hooinClientMutex sync.Mutex
	hooinClient      *commuteclient.HooinCommuteClient

	commuteCtx    context.Context
	cancelCommute context.CancelFunc
	commuteStream *connect.ServerStreamForClient[brainpb.BrainInput]
	commuteDone   chan struct{}

	topicsMutex sync.Mutex
	topics      map[string]*LiveTopic
}

func NewConscious() *Conscious {
	return &Conscious{}
}

func (s *Conscious) Initialize() error {
	return nil
}

func (s *Conscious) checkTopicStarted(topic string) bool {
	s.topicsMutex.Lock()
	defer s.topicsMutex.Unlock()
	if _, ok := s.topics[topic]; ok {
		return true
	}
	return false
}

// ensureTopicStarted must be called from a single goroutine per Conscious instance
// (today: the commute loop). The check-then-act on s.topics is not atomic, so
// concurrent callers for the same new topic would both reach
// Brain.RegisterStepHandler and the second would fail with "step handler
// already registered".
func (s *Conscious) ensureTopicStarted(
	client *commuteclient.HooinCommuteClient, token string, topic string,
) error {
	if s.checkTopicStarted(topic) {
		return nil
	}

	reporter := &StepReporter{client: client, token: token}

	s.topicsMutex.Lock()
	defer s.topicsMutex.Unlock()
	s.topics[topic] = &LiveTopic{reporter: reporter}
	return nil
}

func (s *Conscious) commute(token string) {
	defer func() {
		s.hooinClientMutex.Lock()
		defer s.hooinClientMutex.Unlock()
		s.cancelCommute()
		close(s.commuteDone)
		s.hooinClient = nil
		s.commuteCtx = nil
		s.cancelCommute = nil
		s.commuteStream = nil
		s.commuteDone = nil
	}()
	defer func() {
		s.topicsMutex.Lock()
		defer s.topicsMutex.Unlock()
		wg := sync.WaitGroup{}
		for topic := range s.topics {
			wg.Add(1)
			go func(topic string) {
				defer wg.Done()
				// TODO(nagi): hibernate topic
			}(topic)
		}
		wg.Wait()
		s.topics = nil
	}()
	defer s.commuteStream.Close()

	client := s.hooinClient

	for s.commuteStream.Receive() {
		input := s.commuteStream.Msg()
		err := s.Input(client, token, input)
		if err != nil {
			seedlog.Errorf("Brain input error: %v", err)
		}
	}
	err := s.commuteStream.Err()
	if err != nil {
		// if error is cancel
		if errors.Is(err, context.Canceled) {
			seedlog.Infof("Commute stream closed: %v", err)
		} else {
			seedlog.Errorf("Commute stream error: %v", err)
		}
	}
}

func (s *Conscious) Input(
	client *commuteclient.HooinCommuteClient, token string, input *brainpb.BrainInput,
) error {
	seedlog.Infof("Commute input: %v", input)

	topic := input.GetTopic()
	err := s.ensureTopicStarted(client, token, topic)
	if err != nil {
		return seederr.Wrap(err)
	}

	// Acknowledge the input with an empty result step so the hooin
	// side's SendBrainInput / SendBrainInputStreamBrainStep callers
	// can unblock.
	err = s.hooinClient.ReportBrainStep(
		s.commuteCtx, token, &brainpb.BrainStep{Type: "result"})
	if err != nil {
		return seederr.Wrap(err)
	}
	return nil
}

func (s *Conscious) Wake(
	ctx context.Context,
	token string,
	hooinDirectServer string,
) error {
	s.hooinClientMutex.Lock()
	defer s.hooinClientMutex.Unlock()

	if token == "" {
		return seederr.CodeErrorf(codes.Unauthenticated, "missing token")
	}

	// Wake is exclusive. A Hibernate → Wake sequence can briefly still
	// observe commuteStream != nil: Hibernate only cancels the commute
	// context and returns immediately, while the commute goroutine's
	// cleanup defer (which clears commuteStream under hooinClientMutex)
	// may not have run yet. Callers that need a clean re-wake should
	// wait on Hibernate's returned done channel before calling Wake.
	if s.commuteStream != nil {
		return seederr.CodeErrorf(codes.FailedPrecondition, "already awake")
	}

	if hooinDirectServer == "" {
		return seederr.CodeErrorf(codes.InvalidArgument, "hooin_direct_server is required")
	}
	client := commuteclient.NewHooinCommuteClient(hooinDirectServer)

	commuteCtx, cancelCommute := context.WithCancel(context.Background())
	commuteStream, err := client.Commute(commuteCtx, token)
	if err != nil {
		cancelCommute()
		return seederr.Wrap(err)
	}

	s.hooinClient = client
	s.commuteCtx = commuteCtx
	s.cancelCommute = cancelCommute
	s.commuteStream = commuteStream
	s.commuteDone = make(chan struct{})
	s.topics = map[string]*LiveTopic{}

	go s.commute(token)

	return nil
}

func (s *Conscious) Hibernate() chan struct{} {
	s.hooinClientMutex.Lock()
	defer s.hooinClientMutex.Unlock()
	if s.commuteStream == nil {
		return nil
	}
	s.cancelCommute()
	return s.commuteDone
}
