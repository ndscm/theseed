package brain

import (
	"context"

	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

type BrainStepHandler interface {
	HandleBrainStep(ctx context.Context, topic string, step *brainpb.BrainStep)
}
