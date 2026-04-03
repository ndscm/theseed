package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Stuff holds the schema definition for the Stuff entity.
type Stuff struct {
	ent.Schema
}

// Fields of the Stuff.
func (Stuff) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Unique().Immutable().Default(uuid.New).Annotations(entsql.Default("gen_random_uuid()")),
		field.String("order"),
		field.JSON("data", &json.RawMessage{}).Optional(),
		field.String("owner").Optional().Nillable(),
		field.Time("created_time").Immutable().Default(time.Now),
		field.Time("updated_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Stuff.
func (Stuff) Edges() []ent.Edge {
	return nil
}
