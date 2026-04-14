package tson

import (
	"encoding/json"
	"fmt"
	"strings"
)

// Generate serializes a tson AST node into json.RawMessage.
func Generate(node AstNode) (json.RawMessage, error) {
	b := strings.Builder{}
	err := writeNode(&b, node, 0)
	if err != nil {
		return nil, err
	}
	b.WriteByte('\n')
	return json.RawMessage(b.String()), nil
}

func writeNode(b *strings.Builder, node AstNode, depth int) error {
	switch n := node.(type) {
	case *AstObject:
		return writeObject(b, n, depth)
	case *AstArray:
		return writeArray(b, n, depth)
	case *AstString:
		writeString(b, n.Value)
	case *AstNumber:
		b.WriteString(n.Value)
	case *AstBool:
		if n.Value {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case *AstNull:
		b.WriteString("null")
	default:
		return fmt.Errorf("tson: unknown node type %T", node)
	}
	return nil
}

func writeObject(b *strings.Builder, obj *AstObject, depth int) error {
	if len(obj.Fields) == 0 {
		b.WriteString("{}")
		return nil
	}

	b.WriteString("{\n")
	for i, f := range obj.Fields {
		writeIndent(b, depth+1)
		writeString(b, f.Key)
		b.WriteString(": ")
		err := writeNode(b, f.Value, depth+1)
		if err != nil {
			return err
		}
		if i < len(obj.Fields)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	writeIndent(b, depth)
	b.WriteByte('}')
	return nil
}

func writeArray(b *strings.Builder, arr *AstArray, depth int) error {
	if len(arr.Elements) == 0 {
		b.WriteString("[]")
		return nil
	}

	b.WriteString("[\n")
	for i, elem := range arr.Elements {
		writeIndent(b, depth+1)
		err := writeNode(b, elem, depth+1)
		if err != nil {
			return err
		}
		if i < len(arr.Elements)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	writeIndent(b, depth)
	b.WriteByte(']')
	return nil
}

func writeIndent(b *strings.Builder, depth int) {
	for range depth {
		b.WriteString("  ")
	}
}

func writeString(b *strings.Builder, s string) {
	b.WriteByte('"')
	for _, r := range s {
		switch r {
		case '"':
			b.WriteString(`\"`)
		case '\\':
			b.WriteString(`\\`)
		case '\n':
			b.WriteString(`\n`)
		case '\r':
			b.WriteString(`\r`)
		case '\t':
			b.WriteString(`\t`)
		default:
			b.WriteRune(r)
		}
	}
	b.WriteByte('"')
}
