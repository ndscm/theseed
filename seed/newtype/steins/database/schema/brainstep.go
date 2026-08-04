package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
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
		field.UUID("id", uuid.UUID{}).Unique().Immutable().Default(uuid.New).Annotations(entsql.Default("gen_random_uuid()")),
		field.Time("create_time").Immutable().Default(time.Now),
		field.Time("update_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the BrainStep.
func (BrainStep) Edges() []ent.Edge {
	return nil
}
