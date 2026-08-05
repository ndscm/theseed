package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// BrainStep holds the schema definition for the BrainStep entity.
type BrainStep struct {
	ent.Schema
}

// Fields of the BrainStep.
func (BrainStep) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).StorageKey("uuid").Unique().Immutable().Default(uuid.New),
		field.Time("create_time").Immutable().Default(time.Now),
		field.String("person_id").Immutable(),
		field.Time("emit_time").Immutable(),
		field.String("type").Immutable(),
		field.String("topic").Immutable(),
		field.String("thread_uuid").Immutable(),
		field.JSON("data", map[string]any{}).Optional(),
	}
}

// Edges of the BrainStep.
func (BrainStep) Edges() []ent.Edge {
	return nil
}
