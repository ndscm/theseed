package ast

// AstNode represents a node in the Groovy AST.
//
// The official type name is ASTNode, we keep AstNode here to avoid capitalized abbreviation.
type AstNode interface {
	groovyAstNode()
}

// Statement represents a Groovy statement node.
type Statement interface {
	AstNode
	groovyAstStatement()
}

// Expression represents a Groovy expression node.
type Expression interface {
	AstNode
	groovyAstExpression()
}

// ClassMember represents a member declared inside a class body.
//
// The official ClassNode groups methods, fields, properties, constructors, and
// initializers into separate collections. That is useful for compiler analysis,
// but it loses the original mixed source order across member kinds. The
// formatter needs that order, so class bodies use this single mixed member layer
// instead of grouped member slices.
type ClassMember interface {
	AstNode
	groovyAstClassMember()
}

// ModuleNode is the root AST for one Groovy source file.
type ModuleNode struct {
	Package        *PackageNode
	Imports        []*ImportNode
	Classes        []*ClassNode
	Methods        []*MethodNode
	StatementBlock *BlockStatement
	Statements     []Statement
}

func (ModuleNode) groovyAstNode() {}

// PackageNode represents a package declaration.
type PackageNode struct {
	Name string
}

func (PackageNode) groovyAstNode()      {}
func (PackageNode) groovyAstStatement() {}

// ImportNode represents an import declaration.
type ImportNode struct {
	ClassName   string
	PackageName string
	FieldName   string
	Alias       string
	IsStar      bool
	IsStatic    bool
}

func (ImportNode) groovyAstNode()      {}
func (ImportNode) groovyAstStatement() {}

// AnnotationNode represents an annotation.
type AnnotationNode struct {
	AnnotatedNode

	ClassName string
	Members   []MapEntryExpression
}

func (AnnotationNode) groovyAstNode() {}

// AnnotatedNode carries annotations for declaration nodes.
type AnnotatedNode struct {
	Annotations []AnnotationNode
}

func (AnnotatedNode) groovyAstNode() {}

// ClassNode represents a class, interface, trait, enum, or record declaration.
type ClassNode struct {
	AnnotatedNode

	Kind       string
	Name       string
	Modifiers  []ModifierNode
	Generics   []GenericsType
	SuperClass Expression
	Interfaces []Expression

	// Members preserves the original mixed order of the class body. Keep all
	// formatter-relevant class body declarations here instead of splitting them
	// into methods, fields, properties, and constructors.
	Members []ClassMember
}

func (ClassNode) groovyAstNode()        {}
func (ClassNode) groovyAstStatement()   {}
func (ClassNode) groovyAstClassMember() {}

// InnerClassNode represents an inner class declaration.
type InnerClassNode struct {
	ClassNode
}

// EnumConstantClassNode represents an enum constant class body.
type EnumConstantClassNode struct {
	ClassNode
}

// InterfaceHelperClassNode represents compiler-generated interface helper classes.
type InterfaceHelperClassNode struct {
	ClassNode
}

// ConstructorNode represents a constructor declaration.
type ConstructorNode struct {
	AnnotatedNode

	Modifiers  []ModifierNode
	Parameters []Parameter
	Code       Statement
}

func (ConstructorNode) groovyAstNode()        {}
func (ConstructorNode) groovyAstClassMember() {}

// MethodNode represents a method declaration.
type MethodNode struct {
	AnnotatedNode

	Name       string
	Modifiers  []ModifierNode
	Generics   []GenericsType
	ReturnType Expression
	Parameters []Parameter
	Code       Statement
}

func (MethodNode) groovyAstNode()        {}
func (MethodNode) groovyAstClassMember() {}

// FieldNode represents a field declaration.
type FieldNode struct {
	AnnotatedNode

	Name                   string
	Modifiers              []ModifierNode
	Type                   Expression
	InitialValueExpression Expression
}

func (FieldNode) groovyAstNode()        {}
func (FieldNode) groovyAstClassMember() {}

// PropertyNode represents a property declaration.
type PropertyNode struct {
	AnnotatedNode

	Name                   string
	Modifiers              []ModifierNode
	Type                   Expression
	InitialValueExpression Expression
	GetterBlock            Statement
	SetterBlock            Statement
}

func (PropertyNode) groovyAstNode()        {}
func (PropertyNode) groovyAstClassMember() {}

// RecordComponentNode represents a record component declaration.
type RecordComponentNode struct {
	AnnotatedNode

	Name string
	Type Expression
}

func (RecordComponentNode) groovyAstNode()        {}
func (RecordComponentNode) groovyAstClassMember() {}

// InitializerNode represents an instance or static initializer block.
//
// The official ClassNode stores initializer blocks as statements rather than as
// a dedicated AST node. The formatter keeps them as class members so their
// position relative to fields, methods, constructors, and properties is stable.
type InitializerNode struct {
	Static bool
	Code   Statement
}

func (InitializerNode) groovyAstNode()        {}
func (InitializerNode) groovyAstClassMember() {}

// Parameter represents a method, constructor, closure, or catch parameter.
type Parameter struct {
	Name              string
	Type              Expression
	InitialExpression Expression
}

func (Parameter) groovyAstNode() {}

// GenericsType represents a generic type parameter or type argument.
type GenericsType struct {
	Name       string
	Type       Expression
	LowerBound Expression
	UpperBound []Expression
}

func (GenericsType) groovyAstNode() {}

// ModifierNode represents a declaration modifier.
type ModifierNode struct {
	Name string
}

func (ModifierNode) groovyAstNode() {}

// MixinNode represents a mixin declaration node from the compiler model.
type MixinNode struct {
	Name string
}

func (MixinNode) groovyAstNode() {}

// Variable represents a named variable-like node.
type Variable interface {
	AstNode
	groovyAstVariable()
}

// DynamicVariable represents a dynamically resolved variable.
type DynamicVariable struct {
	Name string
}

func (DynamicVariable) groovyAstNode()     {}
func (DynamicVariable) groovyAstVariable() {}

// VariableScope represents a compiler variable scope.
type VariableScope struct {
}

func (VariableScope) groovyAstNode() {}

// CompileUnit represents a collection of compiled modules.
type CompileUnit struct {
	Modules []ModuleNode
}

func (CompileUnit) groovyAstNode() {}
