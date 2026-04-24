package onduty

import (
	"context"

	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/commute/client/go/commuteclient"
)

type StepReporter struct {
	client *commuteclient.HooinCommuteClient
	token  string
}

func (r *StepReporter) HandleBrainStep(
	ctx context.Context, topic string, step *brainpb.BrainStep,
) {
	err := r.client.ReportBrainStep(ctx, r.token, step)
	if err != nil {
		seedlog.Errorf("Report brain step failed: %v", err)
	}
}
