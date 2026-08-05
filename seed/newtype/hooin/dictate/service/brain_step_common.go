package service

import (
	"github.com/google/uuid"
	"github.com/ndscm/theseed/seed/newtype/gajetto/proto/brainpb"
	"github.com/ndscm/theseed/seed/newtype/steins/database/ent"
	"google.golang.org/protobuf/types/known/structpb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

// getBrainStepProtoFromEnt converts a persisted BrainStep row into its proto
// representation. A row's person_id is server-side metadata with no
// counterpart on the proto, so it is not carried over.
func getBrainStepProtoFromEnt(row *ent.BrainStep) *brainpb.BrainStep {
	if row == nil {
		return nil
	}
	data, err := structpb.NewStruct(row.Data)
	if err != nil {
		data = nil
	}
	brainStep := brainpb.BrainStep{
		Uuid:       row.ID.String(),
		EmitTime:   timestamppb.New(row.EmitTime),
		Type:       row.Type,
		Topic:      row.Topic,
		ThreadUuid: row.ThreadUUID,
		Data:       data,
	}
	return &brainStep
}

// getBrainStepEntFromProto converts a BrainStep proto into an ent row. The
// proto carries no person_id, so PersonID is left unset for the caller to
// populate from request context. A malformed uuid yields the zero id, which
// lets ent generate one on insert.
func getBrainStepEntFromProto(brainStep *brainpb.BrainStep) *ent.BrainStep {
	if brainStep == nil {
		return nil
	}
	row := ent.BrainStep{
		EmitTime:   brainStep.GetEmitTime().AsTime(),
		Type:       brainStep.GetType(),
		Topic:      brainStep.GetTopic(),
		ThreadUUID: brainStep.GetThreadUuid(),
		Data:       brainStep.GetData().AsMap(),
	}
	stepUuid, err := uuid.Parse(brainStep.GetUuid())
	if err == nil {
		row.ID = stepUuid
	}
	return &row
}
