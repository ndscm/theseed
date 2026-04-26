package groovy

import (
	"fmt"
	"strings"
	"unicode"
	"unicode/utf8"
)

func isTrivia(token Token) bool {
	return token.Kind == TokenWhitespace || token.Kind == TokenNewline || token.Kind == TokenComment
}

type lexer struct {
	src    string
	off    int
	line   int
	col    int
	tokens []Token
}

func newLexer(src string) *lexer {
	return &lexer{src: src, line: 1, col: 1}
}

func lex(src string) ([]Token, error) {
	l := newLexer(src)
	for l.off < len(l.src) {
		start := l.position()
		r := l.peek()
		switch {
		case r == '\r' || r == '\n':
			literal := l.scanNewline()
			l.emit(TokenNewline, literal, start)
		case r == ' ' || r == '\t' || r == '\f':
			literal := l.scanWhile(func(r rune) bool { return r == ' ' || r == '\t' || r == '\f' })
			l.emit(TokenWhitespace, literal, start)
		case r == '/' && l.peekN(1) == '/':
			l.emit(TokenComment, l.scanLineComment(), start)
		case r == '/' && l.peekN(1) == '*':
			literal, err := l.scanBlockComment()
			if err != nil {
				return nil, err
			}
			l.emit(TokenComment, literal, start)
		case r == '\'' || r == '"':
			literal, err := l.scanQuotedString(r)
			if err != nil {
				return nil, err
			}
			l.emit(TokenString, literal, start)
		case r == '$' && l.peekN(1) == '/':
			literal, err := l.scanDollarSlashyString()
			if err != nil {
				return nil, err
			}
			l.emit(TokenString, literal, start)
		case unicode.IsLetter(r) || r == '_' || r == '$':
			literal := l.scanIdent()
			kind := TokenIdentifier
			if groovyKeywords[literal] {
				kind = TokenKeyword
			}
			l.emit(kind, literal, start)
		case unicode.IsDigit(r):
			l.emit(TokenNumber, l.scanNumber(), start)
		case r == '/' && l.canStartSlashyString():
			literal, err := l.scanSlashyString()
			if err != nil {
				return nil, err
			}
			l.emit(TokenString, literal, start)
		default:
			l.emit(TokenSymbol, l.scanSymbol(), start)
		}
	}
	l.emit(TokenEOF, "", l.position())
	return l.tokens, nil
}

func (l *lexer) position() Position {
	return Position{Offset: l.off, Line: l.line, Column: l.col}
}

func (l *lexer) emit(kind TokenKind, literal string, start Position) {
	l.tokens = append(l.tokens, Token{Kind: kind, Literal: literal, Range: Range{Start: start, End: l.position()}})
}

func (l *lexer) peek() rune {
	if l.off >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[l.off:])
	return r
}

func (l *lexer) peekN(n int) rune {
	off := l.off
	for i := 0; i < n && off < len(l.src); i++ {
		_, size := utf8.DecodeRuneInString(l.src[off:])
		off += size
	}
	if off >= len(l.src) {
		return 0
	}
	r, _ := utf8.DecodeRuneInString(l.src[off:])
	return r
}

func (l *lexer) advance() rune {
	if l.off >= len(l.src) {
		return 0
	}
	r, size := utf8.DecodeRuneInString(l.src[l.off:])
	l.off += size
	switch r {
	case '\n':
		l.line++
		l.col = 1
	case '\r':
		// \r\n counts as one line break; defer the bump to the following \n.
		if l.off >= len(l.src) || l.src[l.off] != '\n' {
			l.line++
			l.col = 1
		} else {
			l.col++
		}
	default:
		l.col++
	}
	return r
}

func (l *lexer) scanWhile(fn func(rune) bool) string {
	start := l.off
	for l.off < len(l.src) && fn(l.peek()) {
		l.advance()
	}
	return l.src[start:l.off]
}

func (l *lexer) scanNewline() string {
	start := l.off
	if l.peek() == '\r' {
		l.off++
		if l.off < len(l.src) && l.src[l.off] == '\n' {
			l.off++
		}
		l.line++
		l.col = 1
		return l.src[start:l.off]
	}
	l.advance()
	return l.src[start:l.off]
}

func (l *lexer) scanLineComment() string {
	start := l.off
	for l.off < len(l.src) && l.peek() != '\r' && l.peek() != '\n' {
		l.advance()
	}
	return l.src[start:l.off]
}

func (l *lexer) scanBlockComment() (string, error) {
	start := l.off
	l.advance()
	l.advance()
	for l.off < len(l.src) {
		if l.peek() == '*' && l.peekN(1) == '/' {
			l.advance()
			l.advance()
			return l.src[start:l.off], nil
		}
		l.advance()
	}
	return "", fmt.Errorf("unterminated block comment at byte %d", start)
}

func (l *lexer) scanQuotedString(quote rune) (string, error) {
	start := l.off
	triple := l.peekN(1) == quote && l.peekN(2) == quote
	l.advance()
	if triple {
		l.advance()
		l.advance()
	}
	for l.off < len(l.src) {
		if l.peek() == '\\' {
			l.advance()
			if l.off < len(l.src) {
				l.advance()
			}
			continue
		}
		if triple {
			if l.peek() == quote && l.peekN(1) == quote && l.peekN(2) == quote {
				l.advance()
				l.advance()
				l.advance()
				return l.src[start:l.off], nil
			}
		} else if l.peek() == quote {
			l.advance()
			return l.src[start:l.off], nil
		}
		l.advance()
	}
	return "", fmt.Errorf("unterminated string at byte %d", start)
}

func (l *lexer) scanDollarSlashyString() (string, error) {
	start := l.off
	l.advance()
	l.advance()
	for l.off < len(l.src) {
		if l.peek() == '/' && l.peekN(1) == '$' {
			l.advance()
			l.advance()
			return l.src[start:l.off], nil
		}
		l.advance()
	}
	return "", fmt.Errorf("unterminated dollar-slashy string at byte %d", start)
}

func (l *lexer) scanSlashyString() (string, error) {
	start := l.off
	startLine := l.line
	startCol := l.col
	l.advance()
	for l.off < len(l.src) {
		if l.peek() == '\\' {
			l.advance()
			if l.off < len(l.src) {
				l.advance()
			}
			continue
		}
		if l.peek() == '/' {
			l.advance()
			return l.src[start:l.off], nil
		}
		if l.peek() == '\r' || l.peek() == '\n' {
			break
		}
		l.advance()
	}
	l.off = start
	l.line = startLine
	l.col = startCol
	l.advance()
	return l.src[start:l.off], nil
}

func (l *lexer) canStartSlashyString() bool {
	for i := len(l.tokens) - 1; i >= 0; i-- {
		tok := l.tokens[i]
		if isTrivia(tok) {
			continue
		}
		if tok.Kind == TokenKeyword {
			return tok.Literal == "return" || tok.Literal == "throw" || tok.Literal == "case" || tok.Literal == "in"
		}
		if tok.Kind != TokenSymbol {
			return false
		}
		switch tok.Literal {
		case ")", "]", "}", "++", "--":
			return false
		}
		return true
	}
	return true
}

func (l *lexer) scanIdent() string {
	return l.scanWhile(func(r rune) bool {
		return unicode.IsLetter(r) || unicode.IsDigit(r) || r == '_' || r == '$'
	})
}

func (l *lexer) scanNumber() string {
	start := l.off
	for l.off < len(l.src) {
		r := l.peek()
		if unicode.IsDigit(r) || unicode.IsLetter(r) || r == '_' {
			l.advance()
			continue
		}
		if r == '.' && unicode.IsDigit(l.peekN(1)) {
			l.advance()
			continue
		}
		break
	}
	return l.src[start:l.off]
}

func (l *lexer) scanSymbol() string {
	start := l.off
	for _, symbol := range groovySymbols {
		if strings.HasPrefix(l.src[l.off:], symbol) {
			for range symbol {
				l.advance()
			}
			return l.src[start:l.off]
		}
	}
	l.advance()
	return l.src[start:l.off]
}
