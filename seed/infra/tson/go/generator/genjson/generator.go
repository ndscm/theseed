package genjson

import (
	"encoding/json"
	"strings"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/tson/go/tsonast"
)

// Generate serializes a tson AST node into json.RawMessage.
func Generate(node tsonast.AstNode) (json.RawMessage, error) {
	b := strings.Builder{}
	err := writeNode(&b, node, 0)
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	b.WriteByte('\n')
	return json.RawMessage(b.String()), nil
}

func writeNode(b *strings.Builder, node tsonast.AstNode, depth int) error {
	switch n := node.(type) {
	case *tsonast.AstObject:
		return writeObject(b, n, depth)
	case *tsonast.AstArray:
		return writeArray(b, n, depth)
	case *tsonast.AstString:
		writeString(b, n.Value)
	case *tsonast.AstNumber:
		b.WriteString(n.Value)
	case *tsonast.AstBool:
		if n.Value {
			b.WriteString("true")
		} else {
			b.WriteString("false")
		}
	case *tsonast.AstNull:
		b.WriteString("null")
	default:
		return seederr.WrapErrorf("tson: unknown node type %T", node)
	}
	return nil
}

func writeObject(b *strings.Builder, obj *tsonast.AstObject, depth int) error {
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
			return seederr.Wrap(err)
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

func writeArray(b *strings.Builder, arr *tsonast.AstArray, depth int) error {
	if len(arr.Elements) == 0 {
		b.WriteString("[]")
		return nil
	}

	b.WriteString("[\n")
	for i, elem := range arr.Elements {
		writeIndent(b, depth+1)
		err := writeNode(b, elem, depth+1)
		if err != nil {
			return seederr.Wrap(err)
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
	enc, _ := json.Marshal(s)
	b.Write(enc)
}
