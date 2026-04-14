package tson

import "testing"

func TestGenerateObject(t *testing.T) {
	src := `export default {
  name: "Nagi",
  age: 33,
}`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Generate(node)
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "name": "Nagi",
  "age": 33
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

	node, err := Parse(src)
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
	node := &AstObject{}
	got, err := Generate(node)
	if err != nil {
		t.Fatal(err)
	}
	if string(got) != "{}\n" {
		t.Errorf("got %q, want %q", string(got), "{}\n")
	}
}

func TestGenerateStringEscapes(t *testing.T) {
	node := &AstObject{
		Fields: []AstField{
			{Key: "msg", Value: &AstString{Value: "line1\nline2"}},
			{Key: "path", Value: &AstString{Value: `C:\Users`}},
			{Key: "quote", Value: &AstString{Value: `say "hi"`}},
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
  name: string
  age: number
}

export default {
  name: "Nagi",
  age: 33,
} satisfies User`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	got, err := Generate(node)
	if err != nil {
		t.Fatal(err)
	}

	want := `{
  "name": "Nagi",
  "age": 33
}
`
	if string(got) != want {
		t.Errorf("got:\n%s\nwant:\n%s", string(got), want)
	}
}
