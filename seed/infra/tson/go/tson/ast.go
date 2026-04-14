package tson

// AstNode represents a value in the tson AST.
type AstNode interface {
	tsonNode()
}

// AstObject represents a JSON-compatible object literal.
type AstObject struct {
	AstNode
	Fields []AstField
}

// AstField is a key-value pair in an AstObject.
type AstField struct {
	Key   string
	Value AstNode
}

// AstArray represents a JSON-compatible array literal.
type AstArray struct {
	AstNode
	Elements []AstNode
}

// AstString represents a string literal.
type AstString struct {
	AstNode
	Value string
}

// AstNumber represents a number literal, stored as raw text to preserve precision.
type AstNumber struct {
	AstNode
	Value string
}

// AstBool represents a boolean literal.
type AstBool struct {
	AstNode
	Value bool
}

// AstNull represents the null literal.
type AstNull struct {
	AstNode
}

