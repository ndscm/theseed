package tson

import "fmt"

// Parse parses a tson (.config.ts) source string and returns the AST node
// for the export default value.
func Parse(src string) (AstNode, error) {
	p := &parser{lex: newLexer(src)}
	err := p.next()
	if err != nil {
		return nil, err
	}

	for p.tok.kind != tokEOF {
		if p.tok.kind != tokIdent {
			return nil, p.errorf("expected statement, got %q", p.tok.val)
		}
		switch p.tok.val {
		case "import":
			err := p.skipImport()
			if err != nil {
				return nil, err
			}
		case "type":
			err := p.skipTypeDecl()
			if err != nil {
				return nil, err
			}
		case "export":
			return p.parseExportDefault()
		default:
			return nil, p.errorf("unexpected statement %q", p.tok.val)
		}
	}

	return nil, fmt.Errorf("tson: no export default found")
}

type parser struct {
	lex *lexer
	tok token
}

func (p *parser) next() error {
	tok, err := p.lex.next()
	if err != nil {
		return err
	}
	p.tok = tok
	return nil
}

func (p *parser) errorf(format string, args ...any) error {
	prefix := fmt.Sprintf("%d:%d: ", p.tok.line, p.tok.col)
	return fmt.Errorf(prefix+format, args...)
}

func (p *parser) expect(kind tokenKind) error {
	if p.tok.kind != kind {
		return p.errorf("unexpected %q", p.tok.val)
	}
	return p.next()
}

// skipImport consumes an import statement:
//
//	import { type Foo, type Bar } from "./path"
//	import type { Foo } from "./path"
func (p *parser) skipImport() error {
	err := p.next() // skip "import"
	if err != nil {
		return err
	}

	// import type { ... } from "..."
	if p.tok.kind == tokIdent && p.tok.val == "type" {
		err = p.next()
		if err != nil {
			return err
		}
	}

	if p.tok.kind != tokLBrace {
		return p.errorf("expected '{' in import")
	}
	err = p.next()
	if err != nil {
		return err
	}

	// skip until closing brace
	for p.tok.kind != tokRBrace {
		if p.tok.kind == tokEOF {
			return p.errorf("unexpected EOF in import")
		}
		err = p.next()
		if err != nil {
			return err
		}
	}
	err = p.next() // skip "}"
	if err != nil {
		return err
	}

	// "from"
	if p.tok.kind != tokIdent || p.tok.val != "from" {
		return p.errorf("expected 'from'")
	}
	err = p.next()
	if err != nil {
		return err
	}

	// module path string
	if p.tok.kind != tokString {
		return p.errorf("expected module path string")
	}
	err = p.next()
	if err != nil {
		return err
	}

	// optional semicolon
	if p.tok.kind == tokSemicolon {
		err = p.next()
		if err != nil {
			return err
		}
	}

	return nil
}

// skipTypeDecl consumes a type declaration:
//
//	type Foo = { name: string; age: number }
//	type Bar = string | number
func (p *parser) skipTypeDecl() error {
	err := p.next() // skip "type"
	if err != nil {
		return err
	}

	if p.tok.kind != tokIdent {
		return p.errorf("expected type name")
	}
	err = p.next() // skip type name
	if err != nil {
		return err
	}

	if p.tok.kind != tokEquals {
		return p.errorf("expected '=' in type declaration")
	}
	err = p.next() // skip "="
	if err != nil {
		return err
	}

	// Skip the type expression by tracking brace/bracket/paren depth.
	// Stop when at depth 0 and we encounter a statement keyword or EOF.
	depth := 0
	for {
		switch p.tok.kind {
		case tokEOF:
			return nil
		case tokLBrace, tokLBracket, tokLParen:
			depth++
		case tokRBrace, tokRBracket, tokRParen:
			depth--
		case tokSemicolon:
			if depth <= 0 {
				return p.next()
			}
		case tokIdent:
			if depth <= 0 && isStatementKeyword(p.tok.val) {
				return nil
			}
		}
		err = p.next()
		if err != nil {
			return err
		}
	}
}

func isStatementKeyword(s string) bool {
	return s == "export" || s == "import" || s == "type"
}

func (p *parser) parseExportDefault() (AstNode, error) {
	err := p.next() // skip "export"
	if err != nil {
		return nil, err
	}
	if p.tok.kind != tokIdent || p.tok.val != "default" {
		return nil, p.errorf("expected 'default' after 'export'")
	}
	err = p.next() // skip "default"
	if err != nil {
		return nil, err
	}

	node, err := p.parseValue()
	if err != nil {
		return nil, err
	}

	// skip optional "satisfies Type"
	if p.tok.kind == tokIdent && p.tok.val == "satisfies" {
		err = p.next()
		if err != nil {
			return nil, err
		}
		if p.tok.kind != tokIdent {
			return nil, p.errorf("expected type name after 'satisfies'")
		}
		err = p.next()
		if err != nil {
			return nil, err
		}
	}

	return node, nil
}

func (p *parser) parseValue() (AstNode, error) {
	switch p.tok.kind {
	case tokLBrace:
		return p.parseObject()
	case tokLBracket:
		return p.parseArray()
	case tokString:
		val := p.tok.val
		err := p.next()
		if err != nil {
			return nil, err
		}
		return &AstString{Value: val}, nil
	case tokNumber:
		val := p.tok.val
		err := p.next()
		if err != nil {
			return nil, err
		}
		return &AstNumber{Value: val}, nil
	case tokMinus:
		err := p.next()
		if err != nil {
			return nil, err
		}
		if p.tok.kind != tokNumber {
			return nil, p.errorf("expected number after '-'")
		}
		val := "-" + p.tok.val
		err = p.next()
		if err != nil {
			return nil, err
		}
		return &AstNumber{Value: val}, nil
	case tokIdent:
		switch p.tok.val {
		case "true":
			err := p.next()
			if err != nil {
				return nil, err
			}
			return &AstBool{Value: true}, nil
		case "false":
			err := p.next()
			if err != nil {
				return nil, err
			}
			return &AstBool{Value: false}, nil
		case "null":
			err := p.next()
			if err != nil {
				return nil, err
			}
			return &AstNull{}, nil
		default:
			return nil, p.errorf("unexpected identifier %q in value position", p.tok.val)
		}
	default:
		return nil, p.errorf("unexpected %q", p.tok.val)
	}
}

func (p *parser) parseObject() (AstNode, error) {
	err := p.expect(tokLBrace)
	if err != nil {
		return nil, err
	}

	obj := &AstObject{}
	for p.tok.kind != tokRBrace {
		if p.tok.kind == tokEOF {
			return nil, p.errorf("unexpected EOF in object")
		}

		// key: identifier, string, or number
		key := ""
		switch p.tok.kind {
		case tokIdent:
			key = p.tok.val
		case tokString:
			key = p.tok.val
		case tokNumber:
			key = p.tok.val
		default:
			return nil, p.errorf("expected object key, got %q", p.tok.val)
		}
		err = p.next()
		if err != nil {
			return nil, err
		}

		// colon
		if p.tok.kind != tokColon {
			return nil, p.errorf("expected ':' after object key")
		}
		err = p.next()
		if err != nil {
			return nil, err
		}

		// value
		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}

		obj.Fields = append(obj.Fields, AstField{Key: key, Value: val})

		// optional trailing comma
		if p.tok.kind == tokComma {
			err = p.next()
			if err != nil {
				return nil, err
			}
		}
	}

	err = p.next() // skip "}"
	if err != nil {
		return nil, err
	}
	return obj, nil
}

func (p *parser) parseArray() (AstNode, error) {
	err := p.expect(tokLBracket)
	if err != nil {
		return nil, err
	}

	arr := &AstArray{}
	for p.tok.kind != tokRBracket {
		if p.tok.kind == tokEOF {
			return nil, p.errorf("unexpected EOF in array")
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, err
		}
		arr.Elements = append(arr.Elements, val)

		// optional trailing comma
		if p.tok.kind == tokComma {
			err = p.next()
			if err != nil {
				return nil, err
			}
		}
	}

	err = p.next() // skip "]"
	if err != nil {
		return nil, err
	}
	return arr, nil
}
