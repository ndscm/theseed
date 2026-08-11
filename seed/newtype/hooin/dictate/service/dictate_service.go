package service

import (
	"context"
	"time"

	"connectrpc.com/connect"
	"entgo.io/ent/dialect/sql"
	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/cloud/login/go/login"
	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/grpc/go/aips"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepb"
	"github.com/ndscm/theseed/seed/newtype/hooin/dictate/proto/dictatepbconnect"
	"github.com/ndscm/theseed/seed/newtype/hooin/onsite"
	"github.com/ndscm/theseed/seed/newtype/steins/database/ent"
	"github.com/ndscm/theseed/seed/newtype/steins/database/ent/brainstep"
	"github.com/ndscm/theseed/seed/newtype/steins/database/steinsdb"
	"google.golang.org/grpc/codes"
)

func prepareBrainInput(brainInput *brainpb.BrainInput) *brainpb.BrainInput {
	if brainInput.GetUuid() == "" {
		brainInput.Uuid = uuid.NewString()
	}
	if brainInput.GetTopic() == "" {
		brainInput.Topic = "default"
	}
	if brainInput.GetThreadUuid() == "" {
		brainInput.ThreadUuid = brainInput.GetUuid()
	}
	return brainInput
}

type HooinDictateService struct {
	office *onsite.Office
}

func (svc *HooinDictateService) ListBrainSteps(
	ctx context.Context,
	req *connect.Request[dictatepb.ListBrainStepsRequest],
) (*connect.Response[dictatepb.ListBrainStepsResponse], error) {
	_, err := login.EnsureLoginUser(ctx)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	// TODO(nagi): add fine-grained authorization
	parent, err := aips.ParseResourceName(req.Msg.GetParent())
	if err != nil {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "%v", err)
	}
	personId := parent.Ids["people"]
	if personId == "" {
		return nil, seederr.CodeErrorf(codes.InvalidArgument, "parent must have the form %q", "people/{person_id}")
	}
	db := svc.office.GetDatabase()
	if db == nil {
		return nil, seederr.CodeErrorf(codes.FailedPrecondition, "brain step history is not available")
	}

	filters := []*sql.Predicate{sql.EQ(brainstep.FieldPersonID, personId)}
	if req.Msg.GetFilter() != "" {
		filterPred, err := compileBrainStepCelFilter(req.Msg.GetFilter())
		if err != nil {
			return nil, err
		}
		if filterPred != nil {
			filters = append(filters, filterPred)
		}
	}

	orders, err := parseBrainStepOrderBy(req.Msg.GetOrderBy())
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	cursor := (*sql.Predicate)(nil)
	if req.Msg.GetPageToken() != "" {
		cursorId, err := decodeBrainStepPageToken(req.Msg.GetPageToken())
		if err != nil {
			return nil, seederr.CodeErrorf(codes.InvalidArgument, "%v", err)
		}
		// Scope the cursor lookup to personId, not just the raw UUID: an
		// unscoped Get by ID would resolve (and thus reveal the existence
		// of) another person's BrainStep, and would become a real
		// cross-person read seam if the parent scoping above is tightened.
		cursorRow, err := db.BrainStep.Query().
			Where(brainstep.ID(cursorId), brainstep.PersonID(personId)).
			Only(ctx)
		if err != nil {
			return nil, seederr.CodeErrorf(codes.InvalidArgument, "invalid page_token: cursor row not found")
		}
		cursor = buildBrainStepCursorPredicate(cursorRow, orders)
	}

	pageSize := int(req.Msg.GetPageSize())
	if pageSize <= 0 {
		pageSize = defaultPageSize
	} else if pageSize > maxPageSize {
		pageSize = maxPageSize
	}

	rows, totalSize, err := steinsdb.SelectBrainStepRows(ctx, db, filters, orders, cursor, pageSize)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	brainSteps := make([]*brainpb.BrainStep, 0, len(rows))
	for _, row := range rows {
		brainSteps = append(brainSteps, getBrainStepProtoFromEnt(row))
	}

	nextPageToken := ""
	if len(rows) == pageSize {
		nextPageToken = encodeBrainStepPageToken(rows[len(rows)-1].ID)
	}

	response := &dictatepb.ListBrainStepsResponse{
		BrainSteps:    brainSteps,
		NextPageToken: nextPageToken,
		TotalSize:     totalSize,
	}
	return connect.NewResponse(response), nil
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
	brainInput = prepareBrainInput(brainInput)

	sub := onsite.NewStepSubscriber(personId, brainInput.GetTopic(), brainInput.GetThreadUuid())
	svc.office.SubscribeSteps(sub)
	defer svc.office.UnsubscribeSteps(sub)

	err = svc.office.DispatchBrainInput(ctx, personId, brainInput)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	for {
		select {
		case <-ctx.Done():
			return nil, seederr.Wrap(ctx.Err())
		case step := <-sub.Receive():
			if step.GetType() != "result" && step.GetType() != "claudecli-result" {
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
	brainInput = prepareBrainInput(brainInput)

	sub := onsite.NewStepSubscriber(personId, brainInput.GetTopic(), brainInput.GetThreadUuid())
	svc.office.SubscribeSteps(sub)
	defer svc.office.UnsubscribeSteps(sub)

	err = svc.office.DispatchBrainInput(ctx, personId, brainInput)
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
			if step.GetType() == "result" || step.GetType() == "claudecli-result" {
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
		sub := onsite.NewStepSubscriber(pt.GetPersonId(), pt.GetTopic(), "")
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

	// Backfill before draining the live channel. The subscribers above are
	// already registered, so steps emitted during the backfill are buffered and
	// delivered right after; steps that appear in both the persisted backlog and
	// the live buffer are deduplicated on the client by step uuid. If `since` is
	// unset there is nothing to backfill and we serve live steps only.
	//
	// This buffering only spans the bounded channels (`sub.channel` and
	// `merged`, 16 each): while this goroutine is busy in sendBackfillSteps it
	// isn't draining `merged`, so a backfill that outruns those buffers lets
	// BroadcastStep's non-blocking fanout drop further live steps. A dropped
	// step is still persisted asynchronously by saveStep, so the backfill
	// recovers it *if* that write commits before the keyset cursor pages past
	// its emit_time; a tail step persisted after that is missed until the next
	// reconnect. Reconnect gaps are normally small, so this is rare, but it is
	// not the hard "not lost" guarantee a naive reading would assume.
	if since := req.Msg.GetSince(); since != nil {
		err := svc.sendBackfillSteps(ctx, stream, personTopics, since.AsTime())
		if err != nil {
			return err
		}
	}

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

// sendBackfillSteps streams every persisted BrainStep emitted at or after
// `since` that matches any of the subscribed (person, topic) pairs, oldest
// first. It pages through the backlog by keyset so a large gap does not load
// the whole history into memory. An empty person_id or topic on a pair is a
// wildcard, mirroring Office.matchStepSubscribers.
func (svc *HooinDictateService) sendBackfillSteps(
	ctx context.Context,
	stream *connect.ServerStream[brainpb.BrainStep],
	personTopics []*dictatepb.PersonTopic,
	since time.Time,
) error {
	db := svc.office.GetDatabase()
	if db == nil {
		// No history store; nothing to backfill. Live steps still serve.
		return nil
	}

	filters := []*sql.Predicate{sql.GTE(brainstep.FieldEmitTime, since)}
	pairPredicates := make([]*sql.Predicate, 0, len(personTopics))
	matchesAnyPair := false
	for _, pt := range personTopics {
		conditions := make([]*sql.Predicate, 0, 2)
		if pt.GetPersonId() != "" {
			conditions = append(conditions, sql.EQ(brainstep.FieldPersonID, pt.GetPersonId()))
		}
		if pt.GetTopic() != "" {
			conditions = append(conditions, sql.EQ(brainstep.FieldTopic, pt.GetTopic()))
		}
		if len(conditions) == 0 {
			// A fully-wildcard pair matches every step, so the pair
			// restriction adds nothing; drop it entirely.
			matchesAnyPair = true
			break
		}
		pairPredicates = append(pairPredicates, sql.And(conditions...))
	}
	if !matchesAnyPair {
		filters = append(filters, sql.Or(pairPredicates...))
	}

	// Order by (emit_time, id) so the backlog streams in stable chronological
	// order and the keyset cursor below is well defined.
	orders := []aips.FieldOrder{
		{Field: brainstep.FieldEmitTime, Desc: false},
		{Field: brainstep.FieldID, Desc: false},
	}
	var cursorRow *ent.BrainStep
	for {
		cursor := buildBrainStepCursorPredicate(cursorRow, orders)
		rows, _, err := steinsdb.SelectBrainStepRows(ctx, db, filters, orders, cursor, maxPageSize)
		if err != nil {
			return seederr.Wrap(err)
		}
		for _, row := range rows {
			if err := stream.Send(getBrainStepProtoFromEnt(row)); err != nil {
				return seederr.Wrap(err)
			}
		}
		if len(rows) < maxPageSize {
			break
		}
		cursorRow = rows[len(rows)-1]
	}
	return nil
}

var _ dictatepbconnect.HooinDictateServiceHandler = (*HooinDictateService)(nil)

func NewHooinDictateService(office *onsite.Office) *HooinDictateService {
	return &HooinDictateService{
		office: office,
	}
}
