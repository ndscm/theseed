package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/dialect/entsql"
	"entgo.io/ent/schema"
	"entgo.io/ent/schema/field"
)

// Link holds the schema definition for the Link entity.
type Link struct {
	ent.Schema
}

// Annotations of the Golink.
func (Link) Annotations() []schema.Annotation {
	return []schema.Annotation{
		entsql.Annotation{Table: "link"},
	}
}

// Fields of the Link.
func (Link) Fields() []ent.Field {
	return []ent.Field{
		field.String("id").StorageKey("key"),
		field.String("target"),
		field.Bool("public").Default(true),
		field.String("owner").Optional().Nillable(),
		field.Int64("hit_count").Default(0),
		field.Time("created_time").Default(time.Now),
		field.Time("updated_time").Default(time.Now).UpdateDefault(time.Now),
	}
}

// Edges of the Link.
func (Link) Edges() []ent.Edge {
	return nil
}
