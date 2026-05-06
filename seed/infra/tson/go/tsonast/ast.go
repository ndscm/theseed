package tsonast

// AstNode represents a value in the tson AST.
type AstNode interface {
	tsonAstNode()
}

// AstField is a key-value pair in an AstObject.
type AstField struct {
	Key   string
	Value AstNode
}

// AstObject represents a JSON-compatible object literal.
type AstObject struct {
	Fields []AstField
}

func (AstObject) tsonAstNode() {}

// AstArray represents a JSON-compatible array literal.
type AstArray struct {
	Elements []AstNode
}

func (AstArray) tsonAstNode() {}

// AstString represents a string literal.
type AstString struct {
	Value string
}

func (AstString) tsonAstNode() {}

// AstNumber represents a number literal, stored as raw text to preserve precision.
type AstNumber struct {
	Value string
}

func (AstNumber) tsonAstNode() {}

// AstBool represents a boolean literal.
type AstBool struct {
	Value bool
}

func (AstBool) tsonAstNode() {}

// AstNull represents the null literal.
type AstNull struct {
}

func (AstNull) tsonAstNode() {}
