package service

import (
	"context"

	"connectrpc.com/connect"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepb"
	"github.com/ndscm/theseed/seed/newtype/hooin/onsite"
	"google.golang.org/grpc/codes"
)

type HooinDictateService struct {
	office *onsite.Office
}

func (svc *HooinDictateService) Initialize(office *onsite.Office) error {
	svc.office = office
	return nil
}

func (svc *HooinDictateService) SendBrainInput(
	ctx context.Context,
	req *connect.Request[dictatepb.SendBrainInputRequest],
) (*connect.Response[brainpb.BrainStep], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization
	personId := req.Msg.GetPersonId()
	if personId == "" {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "person_id is required")
	}
	brainInput := req.Msg.GetBrainInput()
	if brainInput == nil {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "brain_input is required")
	}

	sub := onsite.NewStepSubscriber(personId, brainInput.GetTopic())
	svc.office.SubscribeSteps(sub)
	defer svc.office.UnsubscribeSteps(sub)

	if err := svc.office.DispatchBrainInput(personId, brainInput); err != nil {
		return nil, seederr.Wrap(err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, seederr.Wrap(ctx.Err())
		case step := <-sub.Receive():
			if step.GetType() != "result" {
				continue
			}
			return connect.NewResponse(step), nil
		}
	}
}

func (svc *HooinDictateService) SendBrainInputStreamBrainStep(
	ctx context.Context,
	req *connect.Request[dictatepb.SendBrainInputRequest],
	stream *connect.ServerStream[brainpb.BrainStep],
) error {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization
	personId := req.Msg.GetPersonId()
	if personId == "" {
		return seederr.CodeErrorf(codes.InvalidArgument, "person_id is required")
	}
	brainInput := req.Msg.GetBrainInput()
	if brainInput == nil {
		return seederr.CodeErrorf(codes.InvalidArgument, "brain_input is required")
	}

	sub := onsite.NewStepSubscriber(personId, brainInput.GetTopic())
	svc.office.SubscribeSteps(sub)
	defer svc.office.UnsubscribeSteps(sub)

	err = svc.office.DispatchBrainInput(personId, brainInput)
	if err != nil {
		return seederr.Wrap(err)
	}

	for {
		select {
		case <-ctx.Done():
			return seederr.Wrap(ctx.Err())
		case step := <-sub.Receive():
			if err := stream.Send(step); err != nil {
				return seederr.Wrap(err)
			}
			if step.GetType() == "result" {
				return nil
			}
		}
	}
}

func (svc *HooinDictateService) SubscribeBrainStep(
	ctx context.Context,
	req *connect.Request[dictatepb.SubscribeBrainStepRequest],
	stream *connect.ServerStream[brainpb.BrainStep],
) error {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization
	personTopics := req.Msg.GetPersonTopics()
	if len(personTopics) == 0 {
		return seederr.CodeErrorf(codes.InvalidArgument, "person_topics is required")
	}

	// Each subscriber owns its own buffered channel (returned via
	// Receive()), so we fan them in onto a single channel that the
	// outbound stream loop drains. The per-sub forwarders exit when
	// ctx is cancelled (i.e. when this RPC returns).
	merged := make(chan *brainpb.BrainStep, 16)
	subs := make([]*onsite.StepSubscriber, 0, len(personTopics))
	for _, pt := range personTopics {
		sub := onsite.NewStepSubscriber(pt.GetPersonId(), pt.GetTopic())
		svc.office.SubscribeSteps(sub)
		subs = append(subs, sub)

		go func(sub *onsite.StepSubscriber) {
			for {
				select {
				case <-ctx.Done():
					return
				case step := <-sub.Receive():
					select {
					case merged <- step:
					case <-ctx.Done():
						return
					}
				}
			}
		}(sub)
	}
	defer func() {
		for _, sub := range subs {
			svc.office.UnsubscribeSteps(sub)
		}
	}()

	for {
		select {
		case <-ctx.Done():
			return seederr.Wrap(ctx.Err())
		case step := <-merged:
			if err := stream.Send(step); err != nil {
				return seederr.Wrap(err)
			}
		}
	}
}
