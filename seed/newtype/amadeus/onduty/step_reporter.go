package onduty

import (
	"context"

	"github.com/ndscm/theseed/seed/cloud/login/go/siliconlogin"
	"github.com/ndscm/theseed/seed/infra/log/go/seedlog"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
)

type StepReporter struct {
	conscious *Conscious
}

func (r *StepReporter) HandleBrainStep(
	ctx context.Context, topic string, step *brainpb.BrainStep,
) {
	client := r.conscious.getCommuteClient()
	if client == nil {
		seedlog.Warnf("Dropping brain step for topic %q: no active connection", topic)
		return
	}
	ctx, err := siliconlogin.SiliconLogin(ctx)
	if err != nil {
		seedlog.Warnf("Silicon login failed: %v", err)
	}
	err = client.ReportBrainStep(ctx, step)
	if err != nil {
		seedlog.Errorf("Report brain step failed: %v", err)
	}
}
