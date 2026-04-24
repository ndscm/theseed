package onsite

import (
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

// StepSubscriber describes a registered listener for BrainSteps.
//
// An empty personId matches any person; an empty topic matches any topic
// for the selected person(s). The owning RPC bounds the subscriber's
// lifetime by calling UnsubscribeSteps on return; broadcasters do not
// consult the subscriber's context directly.
type StepSubscriber struct {
	personId string
	topic    string
	channel  chan *brainpb.BrainStep
}

func NewStepSubscriber(personId string, topic string) *StepSubscriber {
	return &StepSubscriber{
		personId: personId,
		topic:    topic,
		channel:  make(chan *brainpb.BrainStep, 16),
	}
}

func (s *StepSubscriber) Receive() <-chan *brainpb.BrainStep {
	return s.channel
}
