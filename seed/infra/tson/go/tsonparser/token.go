package tsonparser

type tokenKind string

const (
	tokenEOF       tokenKind = "EOF"
	tokenLBrace    tokenKind = "{"
	tokenRBrace    tokenKind = "}"
	tokenLBracket  tokenKind = "["
	tokenRBracket  tokenKind = "]"
	tokenLParen    tokenKind = "("
	tokenRParen    tokenKind = ")"
	tokenColon     tokenKind = ":"
	tokenComma     tokenKind = ","
	tokenSemicolon tokenKind = ";"
	tokenEquals    tokenKind = "="
	tokenMinus     tokenKind = "-"
	tokenString    tokenKind = "String"
	tokenNumber    tokenKind = "Number"
	tokenIdent     tokenKind = "Ident"
	tokenOther     tokenKind = "Other"
)

type token struct {
	kind tokenKind
	val  []byte
	line int
	col  int
}
