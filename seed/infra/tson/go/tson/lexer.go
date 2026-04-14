package tson

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

type tokenKind int

const (
	tokEOF       tokenKind = iota
	tokLBrace              // {
	tokRBrace              // }
	tokLBracket            // [
	tokRBracket            // ]
	tokLParen              // (
	tokRParen              // )
	tokColon               // :
	tokComma               // ,
	tokSemicolon           // ;
	tokEquals              // =
	tokMinus               // -
	tokString              // "..." or '...'
	tokNumber              // 123, 1.5, 1e3
	tokIdent               // keywords and identifiers
	tokOther               // any other single character
)

type token struct {
	kind tokenKind
	val  string
	line int
	col  int
}

type lexer struct {
	src  string
	off  int
	line int
	col  int
}

func newLexer(src string) *lexer {
	return &lexer{src: src, line: 1, col: 1}
}

func (l *lexer) peek() rune {
	if l.off >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.off:])
	return r
}

func (l *lexer) advance() rune {
	if l.off >= len(l.src) {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.src[l.off:])
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

func (l *lexer) next() (token, error) {
	l.skipWhitespaceAndComments()

	if l.off >= len(l.src) {
		return token{kind: tokEOF, line: l.line, col: l.col}, nil
	}

	line, col := l.line, l.col
	r := l.peek()

	switch r {
	case '{':
		l.advance()
		return token{tokLBrace, "{", line, col}, nil
	case '}':
		l.advance()
		return token{tokRBrace, "}", line, col}, nil
	case '[':
		l.advance()
		return token{tokLBracket, "[", line, col}, nil
	case ']':
		l.advance()
		return token{tokRBracket, "]", line, col}, nil
	case '(':
		l.advance()
		return token{tokLParen, "(", line, col}, nil
	case ')':
		l.advance()
		return token{tokRParen, ")", line, col}, nil
	case ':':
		l.advance()
		return token{tokColon, ":", line, col}, nil
	case ',':
		l.advance()
		return token{tokComma, ",", line, col}, nil
	case ';':
		l.advance()
		return token{tokSemicolon, ";", line, col}, nil
	case '=':
		l.advance()
		return token{tokEquals, "=", line, col}, nil
	case '-':
		l.advance()
		return token{tokMinus, "-", line, col}, nil
	case '"', '\'':
		return l.scanString(r, line, col)
	case '`':
		return token{}, fmt.Errorf("%d:%d: template literals are not allowed in tson", line, col)
	default:
		if r >= '0' && r <= '9' {
			return l.scanNumber(line, col), nil
		}
		if r == '_' || r == '$' || unicode.IsLetter(r) {
			return l.scanIdent(line, col), nil
		}
		l.advance()
		return token{tokOther, string(r), line, col}, nil
	}
}

func (l *lexer) scanString(quote rune, line, col int) (token, error) {
	l.advance() // skip opening quote
	b := strings.Builder{}
	for {
		if l.off >= len(l.src) {
			return token{}, fmt.Errorf("%d:%d: unterminated string", line, col)
		}
		r := l.peek()
		if r == '\n' {
			return token{}, fmt.Errorf("%d:%d: unterminated string", line, col)
		}
		if r == quote {
			l.advance()
			return token{tokString, b.String(), line, col}, nil
		}
		if r == '\\' {
			l.advance()
			if l.off >= len(l.src) {
				return token{}, fmt.Errorf("%d:%d: unterminated string", line, col)
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
			case '0':
				b.WriteByte(0)
			case '\'':
				b.WriteByte('\'')
			case '"':
				b.WriteByte('"')
			case 'u':
				if l.off+4 > len(l.src) {
					return token{}, fmt.Errorf("%d:%d: invalid unicode escape", line, col)
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
						return token{}, fmt.Errorf("%d:%d: invalid unicode escape", line, col)
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
	return token{tokNumber, l.src[start:l.off], line, col}
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
	return token{tokIdent, l.src[start:l.off], line, col}
}
