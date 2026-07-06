package brain

import (
	"context"

	"github.com/ndscm/theseed/seed/newtype/amadeus/playpen"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

type Brain interface {
	Initialize(playpenController *playpen.PlaypenController) error
	RegisterStepHandler(topic string, handler BrainStepHandler) error
	Input(ctx context.Context, topic string, input *brainpb.BrainInput) error
	Hibernate(topic string, wait bool) error
}
