package ast

// BlockStatement represents a list of statements.
type BlockStatement struct {
	Statements []Statement
}

func (BlockStatement) groovyAstNode()      {}
func (BlockStatement) groovyAstStatement() {}

// ExpressionStatement represents an expression used as a statement.
type ExpressionStatement struct {
	Expression Expression
}

func (ExpressionStatement) groovyAstNode()      {}
func (ExpressionStatement) groovyAstStatement() {}

// IfStatement represents an if/else statement.
type IfStatement struct {
	BooleanExpression Expression
	IfBlock           Statement
	ElseBlock         Statement
}

func (IfStatement) groovyAstNode()      {}
func (IfStatement) groovyAstStatement() {}

// ForStatement represents a for-in statement.
type ForStatement struct {
	Variable             Parameter
	CollectionExpression Expression
	LoopBlock            Statement
}

func (ForStatement) groovyAstNode()      {}
func (ForStatement) groovyAstStatement() {}
func (ForStatement) groovyAstLooping()   {}

// WhileStatement represents a while statement.
type WhileStatement struct {
	BooleanExpression Expression
	LoopBlock         Statement
}

func (WhileStatement) groovyAstNode()      {}
func (WhileStatement) groovyAstStatement() {}
func (WhileStatement) groovyAstLooping()   {}

// DoWhileStatement represents a do/while statement.
type DoWhileStatement struct {
	BooleanExpression Expression
	LoopBlock         Statement
}

func (DoWhileStatement) groovyAstNode()      {}
func (DoWhileStatement) groovyAstStatement() {}
func (DoWhileStatement) groovyAstLooping()   {}

// LoopingStatement represents a looping statement node.
type LoopingStatement interface {
	Statement
	groovyAstLooping()
}

// ReturnStatement represents a return statement.
type ReturnStatement struct {
	Expression Expression
}

func (ReturnStatement) groovyAstNode()      {}
func (ReturnStatement) groovyAstStatement() {}

// ThrowStatement represents a throw statement.
type ThrowStatement struct {
	Expression Expression
}

func (ThrowStatement) groovyAstNode()      {}
func (ThrowStatement) groovyAstStatement() {}

// TryCatchStatement represents a try/catch/finally statement.
type TryCatchStatement struct {
	TryStatement     Statement
	CatchStatements  []CatchStatement
	FinallyStatement Statement
}

func (TryCatchStatement) groovyAstNode()      {}
func (TryCatchStatement) groovyAstStatement() {}

// CatchStatement represents a catch block.
type CatchStatement struct {
	Variable Parameter
	Code     Statement
}

func (CatchStatement) groovyAstNode()      {}
func (CatchStatement) groovyAstStatement() {}

// SwitchStatement represents a switch statement.
type SwitchStatement struct {
	Expression       Expression
	CaseStatements   []CaseStatement
	DefaultStatement Statement
}

func (SwitchStatement) groovyAstNode()      {}
func (SwitchStatement) groovyAstStatement() {}

// CaseStatement represents a switch case.
type CaseStatement struct {
	Expression Expression
	Code       Statement
}

func (CaseStatement) groovyAstNode()      {}
func (CaseStatement) groovyAstStatement() {}

// BreakStatement represents a break statement.
type BreakStatement struct {
	Label string
}

func (BreakStatement) groovyAstNode()      {}
func (BreakStatement) groovyAstStatement() {}

// ContinueStatement represents a continue statement.
type ContinueStatement struct {
	Label string
}

func (ContinueStatement) groovyAstNode()      {}
func (ContinueStatement) groovyAstStatement() {}

// SynchronizedStatement represents a synchronized block.
type SynchronizedStatement struct {
	Expression Expression
	Code       Statement
}

func (SynchronizedStatement) groovyAstNode()      {}
func (SynchronizedStatement) groovyAstStatement() {}

// AssertStatement represents an assert statement.
type AssertStatement struct {
	BooleanExpression Expression
	MessageExpression Expression
}

func (AssertStatement) groovyAstNode()      {}
func (AssertStatement) groovyAstStatement() {}

// EmptyStatement represents an empty statement.
type EmptyStatement struct {
}

func (EmptyStatement) groovyAstNode()      {}
func (EmptyStatement) groovyAstStatement() {}
