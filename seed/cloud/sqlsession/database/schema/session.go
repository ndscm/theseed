package schema

import (
	"encoding/json"
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Session holds the schema definition for the Session entity.
type Session struct {
	ent.Schema
}

// Annotations of the Session.
func (Session) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "session"},
	}
}

// Fields of the Session.
func (Session) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Unique().Immutable().Default(uuid.New).Annotations(entsql.Default("gen_random_uuid()")),
		field.JSON("data", &json.RawMessage{}).Optional(),
		field.Time("expire_time"),
		field.Time("created_time").Default(time.Now).Immutable(),
		field.Time("updated_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Session.
func (Session) Edges() []ent.Edge {
	return nil
}
