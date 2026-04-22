package claudecli

import (
	"context"

	"github.com/ndscm/theseed/seed/newtype/amadeus/brain"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

type claudeCliBrain struct {
}

func NewClaudeCliBrain() brain.Brain {
	return &claudeCliBrain{}
}

func (b *claudeCliBrain) Initialize() error {
	panic("Initialize is not implemented")
}

func (b *claudeCliBrain) RegisterStepHandler(
	topic string, handler brain.BrainStepHandler,
) error {
	panic("RegisterStepHandler is not implemented")
}

func (b *claudeCliBrain) Input(ctx context.Context, topic string, input *brainpb.BrainInput) error {
	panic("Input is not implemented")
}

func (b *claudeCliBrain) Hibernate(topic string, wait bool) error {
	panic("Hibernate is not implemented")
}
