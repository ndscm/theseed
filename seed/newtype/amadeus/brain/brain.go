package brain

import (
	"context"

	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

type Brain interface {
	Initialize() error
	RegisterStepHandler(topic string, handler BrainStepHandler) error
	Input(ctx context.Context, topic string, input *brainpb.BrainInput) error
	Hibernate(topic string, wait bool) error
}
