package tson

import "testing"

func TestParseSimple(t *testing.T) {
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

	obj, ok := node.(*AstObject)
	if !ok {
		t.Fatalf("expected *Object, got %T", node)
	}
	if len(obj.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(obj.Fields))
	}
	if obj.Fields[0].Key != "name" {
		t.Errorf("field 0 key = %q, want %q", obj.Fields[0].Key, "name")
	}
	s, ok := obj.Fields[0].Value.(*AstString)
	if !ok {
		t.Fatalf("field 0 value: expected *String, got %T", obj.Fields[0].Value)
	}
	if s.Value != "Nagi" {
		t.Errorf("field 0 value = %q, want %q", s.Value, "Nagi")
	}
	if obj.Fields[1].Key != "age" {
		t.Errorf("field 1 key = %q, want %q", obj.Fields[1].Key, "age")
	}
	n, ok := obj.Fields[1].Value.(*AstNumber)
	if !ok {
		t.Fatalf("field 1 value: expected *Number, got %T", obj.Fields[1].Value)
	}
	if n.Value != "33" {
		t.Errorf("field 1 value = %q, want %q", n.Value, "33")
	}
}

func TestParseImport(t *testing.T) {
	src := `import { type User } from "./type/user"

export default {
  name: "Nagi",
  age: 33,
} satisfies User`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	obj, ok := node.(*AstObject)
	if !ok {
		t.Fatalf("expected *Object, got %T", node)
	}
	if len(obj.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(obj.Fields))
	}
}

func TestParseNested(t *testing.T) {
	src := `export default {
  users: [
    { name: "Alice", active: true },
    { name: "Bob", active: false },
  ],
  count: 2,
  metadata: null,
}`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	obj := node.(*AstObject)
	if len(obj.Fields) != 3 {
		t.Fatalf("expected 3 fields, got %d", len(obj.Fields))
	}

	arr, ok := obj.Fields[0].Value.(*AstArray)
	if !ok {
		t.Fatalf("users: expected *Array, got %T", obj.Fields[0].Value)
	}
	if len(arr.Elements) != 2 {
		t.Fatalf("users: expected 2 elements, got %d", len(arr.Elements))
	}

	_, ok = obj.Fields[2].Value.(*AstNull)
	if !ok {
		t.Fatalf("metadata: expected *Null, got %T", obj.Fields[2].Value)
	}
}

func TestParseComments(t *testing.T) {
	src := `// top-level comment
/* block comment */
export default {
  // inline comment
  name: "Nagi", /* trailing comment */
  age: 33,
}`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	obj := node.(*AstObject)
	if len(obj.Fields) != 2 {
		t.Fatalf("expected 2 fields, got %d", len(obj.Fields))
	}
}

func TestParseNegativeNumber(t *testing.T) {
	src := `export default { offset: -10 }`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	obj := node.(*AstObject)
	n := obj.Fields[0].Value.(*AstNumber)
	if n.Value != "-10" {
		t.Errorf("value = %q, want %q", n.Value, "-10")
	}
}

func TestParseStringEscapes(t *testing.T) {
	src := `export default { msg: "hello\nworld", tab: "a\tb" }`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	obj := node.(*AstObject)
	s0 := obj.Fields[0].Value.(*AstString)
	if s0.Value != "hello\nworld" {
		t.Errorf("msg = %q, want %q", s0.Value, "hello\nworld")
	}
	s1 := obj.Fields[1].Value.(*AstString)
	if s1.Value != "a\tb" {
		t.Errorf("tab = %q, want %q", s1.Value, "a\tb")
	}
}

func TestParseSingleQuotedStrings(t *testing.T) {
	src := `export default { name: 'Nagi' }`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	obj := node.(*AstObject)
	s := obj.Fields[0].Value.(*AstString)
	if s.Value != "Nagi" {
		t.Errorf("name = %q, want %q", s.Value, "Nagi")
	}
}

func TestParseNoExportDefault(t *testing.T) {
	src := `type Foo = { name: string }`

	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for missing export default")
	}
}

func TestParseTemplateLiteral(t *testing.T) {
	src := "export default { name: `Nagi` }"

	_, err := Parse(src)
	if err == nil {
		t.Fatal("expected error for template literal")
	}
}

func TestParseImportType(t *testing.T) {
	src := `import type { User } from "./user"

export default {
  name: "Nagi",
}`

	node, err := Parse(src)
	if err != nil {
		t.Fatal(err)
	}

	obj := node.(*AstObject)
	if len(obj.Fields) != 1 {
		t.Fatalf("expected 1 field, got %d", len(obj.Fields))
	}
}
