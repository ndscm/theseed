package onduty

import (
	"context"
	"net/http"
	"sync"

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

type Conscious struct {
	brain brain.Brain

	hooinClientMutex sync.Mutex
	hooinClient      *commuteclient.HooinCommuteClient

	connectHandler http.Handler
	connectCtx     context.Context
	cancelConnect  context.CancelFunc
	connectStream  bidirequest.PayloadStream
	connectDone    chan struct{}

	topicsMutex sync.Mutex
	topics      map[string]*LiveTopic
}

func NewConscious() *Conscious {
	return &Conscious{}
}

func (s *Conscious) Initialize() error {
	b, err := brain.DefaultBrain()
	if err != nil {
		return seederr.Wrap(err)
	}
	s.brain = b
	return nil
}

func (s *Conscious) SetConnectHandler(handler http.Handler) {
	s.connectHandler = handler
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

	reporter := &StepReporter{client: s.hooinClient}
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

func (s *Conscious) commute() {
	defer func() {
		s.hooinClientMutex.Lock()
		defer s.hooinClientMutex.Unlock()
		s.cancelConnect()
		close(s.connectDone)
		s.hooinClient = nil
		s.connectCtx = nil
		s.cancelConnect = nil
		s.connectStream = nil
		s.connectDone = nil
	}()
	defer func() {
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
	defer s.connectStream.Close()

	<-s.connectCtx.Done()

}

func (s *Conscious) Wake() error {
	s.hooinClientMutex.Lock()
	defer s.hooinClientMutex.Unlock()

	// Wake is exclusive. A Hibernate → Wake sequence can briefly still
	// observe connectStream != nil: Hibernate only cancels the connect
	// context and returns immediately, while the commute goroutine
	// (which clears connectStream under hooinClientMutex) may not have
	// run yet. Callers that need a clean re-wake should wait on
	// Hibernate's returned done channel before calling Wake.
	if s.connectStream != nil {
		return seederr.CodeErrorf(codes.FailedPrecondition, "already awake")
	}

	connectClient := bidirequestclient.NewBidirequestClient(commuteclient.HooinCommuteServiceServerFlag())

	connectCtx, cancelConnect := context.WithCancel(context.Background())
	connectCtx, err := siliconlogin.SiliconLogin(connectCtx)
	if err != nil {
		cancelConnect()
		return seederr.Wrap(err)
	}
	connectStream, err := connectClient.Connect(connectCtx)
	if err != nil {
		cancelConnect()
		return seederr.Wrap(err)
	}
	muxClient := bidirequest.WrapClientSide(connectStream, s.connectHandler)

	s.hooinClient = commuteclient.NewHooinCommuteClient(
		commuteclient.WithHttpClient(muxClient),
	)
	s.connectCtx = connectCtx
	s.cancelConnect = cancelConnect
	s.connectStream = connectStream
	s.connectDone = make(chan struct{})
	s.topics = map[string]*LiveTopic{}

	go s.commute()

	return nil
}

func (s *Conscious) Hibernate() chan struct{} {
	s.hooinClientMutex.Lock()
	defer s.hooinClientMutex.Unlock()
	if s.connectStream == nil {
		return nil
	}
	s.cancelConnect()
	return s.connectDone
}
