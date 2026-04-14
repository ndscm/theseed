package genjson

import (
	"testing"

	"github.com/ndscm/theseed/seed/infra/tson/go/tsonast"
	"github.com/ndscm/theseed/seed/infra/tson/go/tsonparser"
)

func TestGenerateObject(t *testing.T) {
	src := `export default {
  id: 1,
  name: "Nagi",
}`

	node, err := tsonparser.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	got, err := Generate(node)
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "id": 1,
  "name": "Nagi"
}
`
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", string(got), want)
	}
}

func TestGenerateNested(t *testing.T) {
	src := `export default {
  users: [
    { name: "Alice", active: true },
    { name: "Bob", active: false },
  ],
  metadata: null,
}`

	node, err := tsonparser.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	got, err := Generate(node)
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "users": [
    {
      "name": "Alice",
      "active": true
    },
    {
      "name": "Bob",
      "active": false
    }
  ],
  "metadata": null
}
`
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", string(got), want)
	}
}

func TestGenerateEmpty(t *testing.T) {
	node := &tsonast.AstObject{}
	got, err := Generate(node)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "{}\n" {
		t.Errorf("got %q, want %q", string(got), "{}\n")
	}
}

func TestGenerateStringEscapes(t *testing.T) {
	node := &tsonast.AstObject{
		Fields: []tsonast.AstField{
			{Key: "msg", Value: &tsonast.AstString{Value: "line1\nline2"}},
			{Key: "path", Value: &tsonast.AstString{Value: `C:\Users`}},
			{Key: "quote", Value: &tsonast.AstString{Value: `say "hi"`}},
		},
	}

	got, err := Generate(node)
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "msg": "line1\nline2",
  "path": "C:\\Users",
  "quote": "say \"hi\""
}
`
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", string(got), want)
	}
}

func TestGenerateRoundTrip(t *testing.T) {
	src := `type User = {
  id: number
  name: string
}

export default {
  id: 1,
  name: "Nagi",
} satisfies User`

	node, err := tsonparser.Parse([]byte(src))
	if err != nil {
		t.Fatal(err)
	}

	got, err := Generate(node)
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "id": 1,
  "name": "Nagi"
}
`
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", string(got), want)
	}
}
