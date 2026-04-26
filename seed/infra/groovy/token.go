package groovy

// TokenKind identifies a lexical token class.
type TokenKind string

const (
	TokenEOF        TokenKind = "EOF"
	TokenWhitespace TokenKind = "Whitespace"
	TokenNewline    TokenKind = "Newline"
	TokenComment    TokenKind = "Comment"
	TokenIdentifier TokenKind = "Identifier"
	TokenKeyword    TokenKind = "Keyword"
	TokenNumber     TokenKind = "Number"
	TokenString     TokenKind = "String"
	TokenSymbol     TokenKind = "Symbol"
)

// Position is a byte offset plus 1-based line and column information.
type Position struct {
	Offset int
	Line   int
	Column int
}

// Range records the source span covered by a token.
type Range struct {
	Start Position
	End   Position
}

// Token is a lexical token used by the parser.
type Token struct {
	Kind    TokenKind
	Literal string
	Range   Range
}
