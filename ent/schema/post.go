package schema

import (
	"time"

	"entgo.io/ent"
	"entgo.io/ent/schema/edge"
	"entgo.io/ent/schema/field"
)

/*
	= POST =
	POST는 제목, 본문, 생성일, 작성자를 포함한다.
	POST는 한 개의 작성자만 가질 수 있다.
*/

// Post holds the schema definition for the Post entity.
type Post struct {
	ent.Schema
}

// Fields of the Post.
func (Post) Fields() []ent.Field {
	return []ent.Field{
		field.String("title"),
		field.Text("body"),
		field.Time("created_at").Default(time.Now),
	}
}

// Edges of the Post.
func (Post) Edges() []ent.Edge {
	return []ent.Edge{
		edge.From("author", User.Type).
			Unique().
			Ref("posts"),
	}
}
