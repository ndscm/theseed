package ast

// ExpressionTransformer represents an expression transformation hook.
type ExpressionTransformer interface {
	TransformExpression(Expression) Expression
}

// VariableExpression represents a variable reference.
type VariableExpression struct {
	Name string
}

func (VariableExpression) groovyAstNode()       {}
func (VariableExpression) groovyAstExpression() {}
func (VariableExpression) groovyAstVariable()   {}

// ConstantExpression represents a scalar constant.
type ConstantExpression struct {
	Kind  string
	Value string
}

func (ConstantExpression) groovyAstNode()       {}
func (ConstantExpression) groovyAstExpression() {}

// TupleExpression represents an ordered expression tuple.
type TupleExpression struct {
	Expressions []Expression
}

func (TupleExpression) groovyAstNode()       {}
func (TupleExpression) groovyAstExpression() {}

// ArgumentListExpression represents method call arguments.
type ArgumentListExpression struct {
	Expressions []Expression
}

func (ArgumentListExpression) groovyAstNode()       {}
func (ArgumentListExpression) groovyAstExpression() {}

// NamedArgumentListExpression represents named arguments in an argument list.
type NamedArgumentListExpression struct {
	MapEntryExpressions []MapEntryExpression
}

func (NamedArgumentListExpression) groovyAstNode()       {}
func (NamedArgumentListExpression) groovyAstExpression() {}

// ListExpression represents a list literal.
type ListExpression struct {
	Expressions []Expression
}

func (ListExpression) groovyAstNode()       {}
func (ListExpression) groovyAstExpression() {}

// MapExpression represents a map literal.
type MapExpression struct {
	Entries []MapEntryExpression
}

func (MapExpression) groovyAstNode()       {}
func (MapExpression) groovyAstExpression() {}

// MapEntryExpression represents one map entry.
type MapEntryExpression struct {
	Key   Expression
	Value Expression
}

func (MapEntryExpression) groovyAstNode()       {}
func (MapEntryExpression) groovyAstExpression() {}

// MethodCall represents an expression that invokes a method-like target.
type MethodCall interface {
	Expression
	groovyAstMethodCall()
}

// MethodCallExpression represents a method call.
type MethodCallExpression struct {
	Object       Expression
	Method       Expression
	Arguments    Expression
	ImplicitThis bool
	Safe         bool
	SpreadSafe   bool
}

func (MethodCallExpression) groovyAstNode()       {}
func (MethodCallExpression) groovyAstExpression() {}
func (MethodCallExpression) groovyAstMethodCall() {}

// StaticMethodCallExpression represents a static method call.
type StaticMethodCallExpression struct {
	OwnerType Expression
	Method    string
	Arguments Expression
}

func (StaticMethodCallExpression) groovyAstNode()       {}
func (StaticMethodCallExpression) groovyAstExpression() {}
func (StaticMethodCallExpression) groovyAstMethodCall() {}

// BinaryExpression represents two expressions joined by an operator.
type BinaryExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (BinaryExpression) groovyAstNode()       {}
func (BinaryExpression) groovyAstExpression() {}

// DeclarationExpression represents a variable declaration expression.
type DeclarationExpression struct {
	Left     Expression
	Operator string
	Right    Expression
}

func (DeclarationExpression) groovyAstNode()       {}
func (DeclarationExpression) groovyAstExpression() {}

// PropertyExpression represents property access.
type PropertyExpression struct {
	ObjectExpression Expression
	Property         Expression
	Safe             bool
	SpreadSafe       bool
	ImplicitThis     bool
}

func (PropertyExpression) groovyAstNode()       {}
func (PropertyExpression) groovyAstExpression() {}

// AttributeExpression represents direct field-style attribute access.
type AttributeExpression struct {
	ObjectExpression Expression
	Property         Expression
	Safe             bool
	SpreadSafe       bool
	ImplicitThis     bool
}

func (AttributeExpression) groovyAstNode()       {}
func (AttributeExpression) groovyAstExpression() {}

// FieldExpression represents a field reference expression.
type FieldExpression struct {
	FieldName string
}

func (FieldExpression) groovyAstNode()       {}
func (FieldExpression) groovyAstExpression() {}

// ClosureExpression represents a closure literal.
type ClosureExpression struct {
	Parameters []Parameter
	Code       Statement
}

func (ClosureExpression) groovyAstNode()       {}
func (ClosureExpression) groovyAstExpression() {}

// LambdaExpression represents a lambda expression.
type LambdaExpression struct {
	Parameters []Parameter
	Code       Statement
}

func (LambdaExpression) groovyAstNode()       {}
func (LambdaExpression) groovyAstExpression() {}

// GStringExpression represents an interpolated string expression.
type GStringExpression struct {
	Strings []ConstantExpression
	Values  []Expression
}

func (GStringExpression) groovyAstNode()       {}
func (GStringExpression) groovyAstExpression() {}

// RangeExpression represents a range expression.
type RangeExpression struct {
	From      Expression
	To        Expression
	Inclusive bool
}

func (RangeExpression) groovyAstNode()       {}
func (RangeExpression) groovyAstExpression() {}

// TernaryExpression represents a conditional expression.
type TernaryExpression struct {
	BooleanExpression Expression
	TrueExpression    Expression
	FalseExpression   Expression
}

func (TernaryExpression) groovyAstNode()       {}
func (TernaryExpression) groovyAstExpression() {}

// ElvisOperatorExpression represents an Elvis operator expression.
type ElvisOperatorExpression struct {
	Expression      Expression
	FalseExpression Expression
}

func (ElvisOperatorExpression) groovyAstNode()       {}
func (ElvisOperatorExpression) groovyAstExpression() {}

// BooleanExpression wraps an expression used as a boolean condition.
type BooleanExpression struct {
	Expression Expression
}

func (BooleanExpression) groovyAstNode()       {}
func (BooleanExpression) groovyAstExpression() {}

// ClassExpression represents a class literal expression.
type ClassExpression struct {
	Type Expression
}

func (ClassExpression) groovyAstNode()       {}
func (ClassExpression) groovyAstExpression() {}

// CastExpression represents a cast expression.
type CastExpression struct {
	Type       Expression
	Expression Expression
	Coerce     bool
	Strict     bool
}

func (CastExpression) groovyAstNode()       {}
func (CastExpression) groovyAstExpression() {}

// ConstructorCallExpression represents object construction.
type ConstructorCallExpression struct {
	Type        Expression
	Arguments   Expression
	SpecialCall bool
}

func (ConstructorCallExpression) groovyAstNode()       {}
func (ConstructorCallExpression) groovyAstExpression() {}

// ArrayExpression represents an array literal or array construction.
type ArrayExpression struct {
	ElementType     Expression
	Expressions     []Expression
	SizeExpressions []Expression
}

func (ArrayExpression) groovyAstNode()       {}
func (ArrayExpression) groovyAstExpression() {}

// MethodPointerExpression represents a method pointer expression.
type MethodPointerExpression struct {
	Expression Expression
	MethodName Expression
}

func (MethodPointerExpression) groovyAstNode()       {}
func (MethodPointerExpression) groovyAstExpression() {}

// MethodReferenceExpression represents a method reference expression.
type MethodReferenceExpression struct {
	Expression Expression
	MethodName Expression
}

func (MethodReferenceExpression) groovyAstNode()       {}
func (MethodReferenceExpression) groovyAstExpression() {}

// ClosureListExpression represents a closure list expression.
type ClosureListExpression struct {
	Expressions []Expression
}

func (ClosureListExpression) groovyAstNode()       {}
func (ClosureListExpression) groovyAstExpression() {}

// PrefixExpression represents a prefix operator expression.
type PrefixExpression struct {
	Operator   string
	Expression Expression
}

func (PrefixExpression) groovyAstNode()       {}
func (PrefixExpression) groovyAstExpression() {}

// PostfixExpression represents a postfix operator expression.
type PostfixExpression struct {
	Operator   string
	Expression Expression
}

func (PostfixExpression) groovyAstNode()       {}
func (PostfixExpression) groovyAstExpression() {}

// NotExpression represents a logical not expression.
type NotExpression struct {
	Expression Expression
}

func (NotExpression) groovyAstNode()       {}
func (NotExpression) groovyAstExpression() {}

// UnaryPlusExpression represents a unary plus expression.
type UnaryPlusExpression struct {
	Expression Expression
}

func (UnaryPlusExpression) groovyAstNode()       {}
func (UnaryPlusExpression) groovyAstExpression() {}

// UnaryMinusExpression represents a unary minus expression.
type UnaryMinusExpression struct {
	Expression Expression
}

func (UnaryMinusExpression) groovyAstNode()       {}
func (UnaryMinusExpression) groovyAstExpression() {}

// BitwiseNegationExpression represents a bitwise negation expression.
type BitwiseNegationExpression struct {
	Expression Expression
}

func (BitwiseNegationExpression) groovyAstNode()       {}
func (BitwiseNegationExpression) groovyAstExpression() {}

// SpreadExpression represents a spread expression.
type SpreadExpression struct {
	Expression Expression
}

func (SpreadExpression) groovyAstNode()       {}
func (SpreadExpression) groovyAstExpression() {}

// SpreadMapExpression represents a spread-map expression.
type SpreadMapExpression struct {
	Expression Expression
}

func (SpreadMapExpression) groovyAstNode()       {}
func (SpreadMapExpression) groovyAstExpression() {}

// AnnotationConstantExpression represents an annotation used as a constant expression.
type AnnotationConstantExpression struct {
	Value AnnotationNode
}

func (AnnotationConstantExpression) groovyAstNode()       {}
func (AnnotationConstantExpression) groovyAstExpression() {}

// EmptyExpression represents an empty expression.
type EmptyExpression struct {
}

func (EmptyExpression) groovyAstNode()       {}
func (EmptyExpression) groovyAstExpression() {}
