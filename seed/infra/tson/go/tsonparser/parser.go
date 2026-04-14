package tsonparser

import (
	"strconv"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
	"github.com/ndscm/theseed/seed/infra/tson/go/tsonast"
)

func isStatementKeyword(s []byte) bool {
	v := string(s)
	return v == "export" || v == "import" || v == "type"
}

type parser struct {
	lex *lexer
	tok token
}

func (p *parser) wrapErrorf(format string, args ...any) error {
	prefix := strconv.Itoa(p.tok.line) + ":" + strconv.Itoa(p.tok.col) + ": "
	return seederr.WrapErrorf(prefix+format, args...)
}

func (p *parser) next() error {
	tok, err := p.lex.next()
	if err != nil {
		return seederr.Wrap(err)
	}
	p.tok = tok
	return nil
}

// skipImport consumes an import statement:
//
//	import { type Foo, type Bar } from "./path"
//	import type { Foo } from "./path"
func (p *parser) skipImport() error {
	err := p.next() // skip "import"
	if err != nil {
		return seederr.Wrap(err)
	}

	// import type { ... } from "..."
	if p.tok.kind == tokenIdent && string(p.tok.val) == "type" {
		err = p.next()
		if err != nil {
			return seederr.Wrap(err)
		}
	}

	if p.tok.kind != tokenLBrace {
		return p.wrapErrorf("expected '{' in import")
	}
	err = p.next()
	if err != nil {
		return seederr.Wrap(err)
	}

	// skip until closing brace
	for p.tok.kind != tokenRBrace {
		if p.tok.kind == tokenEOF {
			return p.wrapErrorf("unexpected EOF in import")
		}
		err = p.next()
		if err != nil {
			return seederr.Wrap(err)
		}
	}
	err = p.next() // skip "}"
	if err != nil {
		return seederr.Wrap(err)
	}

	// "from"
	if p.tok.kind != tokenIdent || string(p.tok.val) != "from" {
		return p.wrapErrorf("expected 'from'")
	}
	err = p.next()
	if err != nil {
		return seederr.Wrap(err)
	}

	// module path string
	if p.tok.kind != tokenString {
		return p.wrapErrorf("expected module path string")
	}
	err = p.next()
	if err != nil {
		return seederr.Wrap(err)
	}

	// optional semicolon
	if p.tok.kind == tokenSemicolon {
		err = p.next()
		if err != nil {
			return seederr.Wrap(err)
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
		return seederr.Wrap(err)
	}

	if p.tok.kind != tokenIdent {
		return p.wrapErrorf("expected type name")
	}
	err = p.next() // skip type name
	if err != nil {
		return seederr.Wrap(err)
	}

	if p.tok.kind != tokenEquals {
		return p.wrapErrorf("expected '=' in type declaration")
	}
	err = p.next() // skip "="
	if err != nil {
		return seederr.Wrap(err)
	}

	// Skip the type expression by tracking brace/bracket/paren depth.
	// Stop when at depth 0 and we encounter a statement keyword or EOF.
	depth := 0
	for {
		switch p.tok.kind {
		case tokenEOF:
			return nil
		case tokenLBrace, tokenLBracket, tokenLParen:
			depth++
		case tokenRBrace, tokenRBracket, tokenRParen:
			depth--
		case tokenSemicolon:
			if depth <= 0 {
				return p.next()
			}
		case tokenIdent:
			if depth <= 0 && isStatementKeyword(p.tok.val) {
				return nil
			}
		}
		err = p.next()
		if err != nil {
			return seederr.Wrap(err)
		}
	}
}

func (p *parser) expect(kind tokenKind) error {
	if p.tok.kind != kind {
		return p.wrapErrorf("unexpected %q", p.tok.val)
	}
	return p.next()
}

func (p *parser) parseObject() (tsonast.AstNode, error) {
	err := p.expect(tokenLBrace)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	obj := &tsonast.AstObject{}
	for p.tok.kind != tokenRBrace {
		if p.tok.kind == tokenEOF {
			return nil, p.wrapErrorf("unexpected EOF in object")
		}

		// key: identifier, string, or number
		key := ""
		switch p.tok.kind {
		case tokenIdent:
			key = string(p.tok.val)
		case tokenString:
			key = string(p.tok.val)
		case tokenNumber:
			key = string(p.tok.val)
		default:
			return nil, p.wrapErrorf("expected object key, got %q", p.tok.val)
		}
		err = p.next()
		if err != nil {
			return nil, seederr.Wrap(err)
		}

		// colon
		if p.tok.kind != tokenColon {
			return nil, p.wrapErrorf("expected ':' after object key")
		}
		err = p.next()
		if err != nil {
			return nil, seederr.Wrap(err)
		}

		// value
		val, err := p.parseValue()
		if err != nil {
			return nil, seederr.Wrap(err)
		}

		obj.Fields = append(obj.Fields, tsonast.AstField{Key: key, Value: val})

		// optional trailing comma
		if p.tok.kind == tokenComma {
			err = p.next()
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		}
	}

	err = p.next() // skip "}"
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return obj, nil
}

func (p *parser) parseArray() (tsonast.AstNode, error) {
	err := p.expect(tokenLBracket)
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	arr := &tsonast.AstArray{}
	for p.tok.kind != tokenRBracket {
		if p.tok.kind == tokenEOF {
			return nil, p.wrapErrorf("unexpected EOF in array")
		}

		val, err := p.parseValue()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		arr.Elements = append(arr.Elements, val)

		// optional trailing comma
		if p.tok.kind == tokenComma {
			err = p.next()
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		}
	}

	err = p.next() // skip "]"
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	return arr, nil
}

func (p *parser) parseValue() (tsonast.AstNode, error) {
	switch p.tok.kind {
	case tokenLBrace:
		return p.parseObject()
	case tokenLBracket:
		return p.parseArray()
	case tokenString:
		val := string(p.tok.val)
		err := p.next()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		return &tsonast.AstString{Value: val}, nil
	case tokenNumber:
		val := string(p.tok.val)
		err := p.next()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		return &tsonast.AstNumber{Value: val}, nil
	case tokenMinus:
		err := p.next()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if p.tok.kind != tokenNumber {
			return nil, p.wrapErrorf("expected number after '-'")
		}
		val := "-" + string(p.tok.val)
		err = p.next()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		return &tsonast.AstNumber{Value: val}, nil
	case tokenIdent:
		switch string(p.tok.val) {
		case "true":
			err := p.next()
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			return &tsonast.AstBool{Value: true}, nil
		case "false":
			err := p.next()
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			return &tsonast.AstBool{Value: false}, nil
		case "null":
			err := p.next()
			if err != nil {
				return nil, seederr.Wrap(err)
			}
			return &tsonast.AstNull{}, nil
		default:
			return nil, p.wrapErrorf("unexpected identifier %q in value position", p.tok.val)
		}
	default:
		return nil, p.wrapErrorf("unexpected %q", p.tok.val)
	}
}

func (p *parser) parseExportDefault() (tsonast.AstNode, error) {
	err := p.next() // skip "export"
	if err != nil {
		return nil, seederr.Wrap(err)
	}
	if p.tok.kind != tokenIdent || string(p.tok.val) != "default" {
		return nil, p.wrapErrorf("expected 'default' after 'export'")
	}
	err = p.next() // skip "default"
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	node, err := p.parseValue()
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	// skip optional "satisfies Type"
	if p.tok.kind == tokenIdent && string(p.tok.val) == "satisfies" {
		err = p.next()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
		if p.tok.kind != tokenIdent {
			return nil, p.wrapErrorf("expected type name after 'satisfies'")
		}
		err = p.next()
		if err != nil {
			return nil, seederr.Wrap(err)
		}
	}

	return node, nil
}

// Parse parses a tson (.config.ts) source string and returns the AST node
// for the export default value.
func Parse(src []byte) (tsonast.AstNode, error) {
	p := &parser{lex: newLexer(src)}
	err := p.next()
	if err != nil {
		return nil, seederr.Wrap(err)
	}

	for p.tok.kind != tokenEOF {
		if p.tok.kind != tokenIdent {
			return nil, p.wrapErrorf("expected statement, got %q", p.tok.val)
		}
		switch string(p.tok.val) {
		case "import":
			err := p.skipImport()
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		case "type":
			err := p.skipTypeDecl()
			if err != nil {
				return nil, seederr.Wrap(err)
			}
		case "export":
			return p.parseExportDefault()
		default:
			return nil, p.wrapErrorf("unexpected statement %q", p.tok.val)
		}
	}

	return nil, seederr.WrapErrorf("tson: no export default found")
}
