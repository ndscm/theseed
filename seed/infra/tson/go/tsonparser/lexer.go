package tsonparser

import (
	"bytes"
	"unicode"
	"unicode/utf8"

	"github.com/ndscm/theseed/seed/infra/error/go/seederr"
)

type lexer struct {
	src  []byte
	off  int
	line int
	col  int
}

func (l *lexer) peek() rune {
	if l.off >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRune(l.src[l.off:])
	return r
}

func (l *lexer) advance() rune {
	if l.off >= len(l.src) {
		return 0
	}
	r, size := utf8.DecodeRune(l.src[l.off:])
	l.off += size
	if r == '\n' {
		l.line++
		l.col = 1
	} else {
		l.col++
	}
	return r
}

func (l *lexer) skipWhitespaceAndComments() {
	for l.off < len(l.src) {
		r := l.peek()
		if r == ' ' || r == '\t' || r == '\n' || r == '\r' {
			l.advance()
			continue
		}
		if r == '/' && l.off+1 < len(l.src) {
			switch l.src[l.off+1] {
			case '/':
				l.advance()
				l.advance()
				for l.off < len(l.src) && l.peek() != '\n' {
					l.advance()
				}
				continue
			case '*':
				l.advance()
				l.advance()
				for l.off < len(l.src) {
					if l.peek() == '*' && l.off+1 < len(l.src) && l.src[l.off+1] == '/' {
						l.advance()
						l.advance()
						break
					}
					l.advance()
				}
				continue
			}
		}
		break
	}
}

func (l *lexer) scanString(quote rune, line, col int) (token, error) {
	l.advance() // skip opening quote
	b := bytes.Buffer{}
	for {
		if l.off >= len(l.src) {
			return token{}, seederr.WrapErrorf("%d:%d: unterminated string", line, col)
		}
		r := l.peek()
		if r == '\n' {
			return token{}, seederr.WrapErrorf("%d:%d: unterminated string", line, col)
		}
		if r == quote {
			l.advance()
			return token{tokenString, b.Bytes(), line, col}, nil
		}
		if r == '\\' {
			l.advance()
			if l.off >= len(l.src) {
				return token{}, seederr.WrapErrorf("%d:%d: unterminated string", line, col)
			}
			esc := l.advance()
			switch esc {
			case '\\':
				b.WriteByte('\\')
			case 'n':
				b.WriteByte('\n')
			case 't':
				b.WriteByte('\t')
			case 'r':
				b.WriteByte('\r')
			case 'b':
				b.WriteByte('\b')
			case 'f':
				b.WriteByte('\f')
			case '0':
				b.WriteByte(0)
			case '\'':
				b.WriteByte('\'')
			case '"':
				b.WriteByte('"')
			case 'u':
				if l.off+4 > len(l.src) {
					return token{}, seederr.WrapErrorf("%d:%d: invalid unicode escape", line, col)
				}
				hex := l.src[l.off : l.off+4]
				code := rune(0)
				for _, h := range hex {
					code <<= 4
					switch {
					case h >= '0' && h <= '9':
						code |= rune(h - '0')
					case h >= 'a' && h <= 'f':
						code |= rune(h - 'a' + 10)
					case h >= 'A' && h <= 'F':
						code |= rune(h - 'A' + 10)
					default:
						return token{}, seederr.WrapErrorf("%d:%d: invalid unicode escape", line, col)
					}
				}
				b.WriteRune(code)
				for range 4 {
					l.advance()
				}
			default:
				b.WriteByte('\\')
				b.WriteRune(esc)
			}
		} else {
			b.WriteRune(r)
			l.advance()
		}
	}
}

func (l *lexer) scanNumber(line, col int) token {
	start := l.off
	for l.off < len(l.src) && l.src[l.off] >= '0' && l.src[l.off] <= '9' {
		l.advance()
	}
	if l.off < len(l.src) && l.src[l.off] == '.' {
		l.advance()
		for l.off < len(l.src) && l.src[l.off] >= '0' && l.src[l.off] <= '9' {
			l.advance()
		}
	}
	if l.off < len(l.src) && (l.src[l.off] == 'e' || l.src[l.off] == 'E') {
		l.advance()
		if l.off < len(l.src) && (l.src[l.off] == '+' || l.src[l.off] == '-') {
			l.advance()
		}
		for l.off < len(l.src) && l.src[l.off] >= '0' && l.src[l.off] <= '9' {
			l.advance()
		}
	}
	return token{tokenNumber, l.src[start:l.off], line, col}
}

func (l *lexer) scanIdent(line, col int) token {
	start := l.off
	for l.off < len(l.src) {
		r := l.peek()
		if r == '_' || r == '$' || unicode.IsLetter(r) || unicode.IsDigit(r) {
			l.advance()
		} else {
			break
		}
	}
	return token{tokenIdent, l.src[start:l.off], line, col}
}

func (l *lexer) next() (token, error) {
	l.skipWhitespaceAndComments()

	if l.off >= len(l.src) {
		return token{kind: tokenEOF, line: l.line, col: l.col}, nil
	}

	line, col := l.line, l.col
	r := l.peek()

	switch r {
	case '{':
		l.advance()
		return token{tokenLBrace, []byte("{"), line, col}, nil
	case '}':
		l.advance()
		return token{tokenRBrace, []byte("}"), line, col}, nil
	case '[':
		l.advance()
		return token{tokenLBracket, []byte("["), line, col}, nil
	case ']':
		l.advance()
		return token{tokenRBracket, []byte("]"), line, col}, nil
	case '(':
		l.advance()
		return token{tokenLParen, []byte("("), line, col}, nil
	case ')':
		l.advance()
		return token{tokenRParen, []byte(")"), line, col}, nil
	case ':':
		l.advance()
		return token{tokenColon, []byte(":"), line, col}, nil
	case ',':
		l.advance()
		return token{tokenComma, []byte(","), line, col}, nil
	case ';':
		l.advance()
		return token{tokenSemicolon, []byte(";"), line, col}, nil
	case '=':
		l.advance()
		return token{tokenEquals, []byte("="), line, col}, nil
	case '-':
		l.advance()
		return token{tokenMinus, []byte("-"), line, col}, nil
	case '"', '\'':
		return l.scanString(r, line, col)
	case '`':
		return token{}, seederr.WrapErrorf("%d:%d: template literals are not allowed in tson", line, col)
	default:
		if r >= '0' && r <= '9' {
			return l.scanNumber(line, col), nil
		}
		if r == '_' || r == '$' || unicode.IsLetter(r) {
			return l.scanIdent(line, col), nil
		}
		l.advance()
		return token{tokenOther, []byte(string(r)), line, col}, nil
	}
}

func newLexer(src []byte) *lexer {
	return &lexer{src: src, line: 1, col: 1}
}
