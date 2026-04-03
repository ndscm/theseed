package schema

import (
	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema/field"
	"github.com/google/uuid"
)

// Category holds the schema definition for the Category entity.
type Category struct {
	ent.Schema
}

// Fields of the Category.
func (Category) Fields() []ent.Field {
	return []ent.Field{
		field.UUID("id", uuid.UUID{}).Unique().Immutable().Default(uuid.New).Annotations(entsql.Default("gen_random_uuid()")),
		field.String("primary").Optional().Nillable(),
		field.String("secondary").Optional().Nillable(),
	}
}

// Edges of the Category.
func (Category) Edges() []ent.Edge {
	return nil
}
