package groovy

import (
	"fmt"
	"strings"

	"github.com/ndscm/theseed/seed/infra/groovy/ast"
)

type directiveKind int

const (
	directivePackage directiveKind = iota
	directiveImport
)

func skipNewlines(tokens []Token, index int) int {
	for index < len(tokens) && tokens[index].Kind == TokenNewline {
		index++
	}
	return index
}

func trimExpressionTokens(tokens []Token) []Token {
	start := 0
	for start < len(tokens) && tokens[start].Kind == TokenNewline {
		start++
	}
	end := len(tokens)
	for end > start && (tokens[end-1].Kind == TokenNewline || tokens[end-1].Kind == TokenEOF) {
		end--
	}
	return tokens[start:end]
}

func trimStatementTokens(tokens []Token) []Token {
	tokens = trimExpressionTokens(tokens)
	if len(tokens) > 0 && tokens[len(tokens)-1].Literal == ";" {
		tokens = tokens[:len(tokens)-1]
	}
	return trimExpressionTokens(tokens)
}

func literalForTokens(tokens []Token) string {
	b := strings.Builder{}
	for _, token := range tokens {
		if token.Kind == TokenEOF || token.Kind == TokenNewline {
			continue
		}
		b.WriteString(token.Literal)
	}
	return b.String()
}

func splitTopLevel(tokens []Token, literal string) [][]Token {
	var parts [][]Token
	start := 0
	depth := 0
	for i, token := range tokens {
		if token.Kind == TokenSymbol {
			switch token.Literal {
			case "(", "[", "{":
				depth++
			case ")", "]", "}":
				if depth > 0 {
					depth--
				}
			}
		}
		if depth == 0 && token.Literal == literal {
			parts = append(parts, tokens[start:i])
			start = i + 1
		}
	}
	parts = append(parts, tokens[start:])
	return parts
}

func topLevelLiteralIndex(tokens []Token, literal string) int {
	depth := 0
	for i, token := range tokens {
		if token.Kind == TokenSymbol {
			switch token.Literal {
			case "(", "[", "{":
				depth++
			case ")", "]", "}":
				if depth > 0 {
					depth--
				}
			}
		}
		if depth == 0 && token.Literal == literal {
			return i
		}
	}
	return -1
}

func topLevelOpenLiteralIndex(tokens []Token, literal string) int {
	depth := 0
	for i, token := range tokens {
		if depth == 0 && token.Literal == literal {
			return i
		}
		if token.Kind != TokenSymbol {
			continue
		}
		switch token.Literal {
		case "(", "[", "{":
			depth++
		case ")", "]", "}":
			if depth > 0 {
				depth--
			}
		}
	}
	return -1
}

func enclosingPair(tokens []Token, open string, close string) (int, int) {
	openIndex := -1
	depth := 0
	for i, token := range tokens {
		if token.Kind != TokenSymbol {
			continue
		}
		switch token.Literal {
		case open:
			if depth == 0 && openIndex < 0 {
				openIndex = i
			}
			depth++
		case close:
			if depth > 0 {
				depth--
			}
			if depth == 0 && openIndex >= 0 {
				return openIndex, i
			}
		}
	}
	return -1, -1
}

func matchingClose(tokens []Token, openIndex int, open string, close string) int {
	depth := 0
	for i := openIndex; i < len(tokens); i++ {
		token := tokens[i]
		if token.Kind != TokenSymbol {
			continue
		}
		switch token.Literal {
		case open:
			depth++
		case close:
			depth--
			if depth == 0 {
				return i
			}
		}
	}
	return -1
}

func firstIdentifierAfter(tokens []Token, start int) string {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Kind == TokenIdentifier {
			return tokens[i].Literal
		}
	}
	return ""
}

func firstNonNewlineLiteral(tokens []Token) string {
	for _, token := range tokens {
		if token.Kind == TokenNewline {
			continue
		}
		return token.Literal
	}
	return ""
}

func firstIdentifierIndexAfter(tokens []Token, start int) int {
	for i := start; i < len(tokens); i++ {
		if tokens[i].Kind == TokenIdentifier {
			return i
		}
	}
	return -1
}

func skipClassBodyTrivia(tokens []Token, index int) int {
	for index < len(tokens) && (tokens[index].Kind == TokenNewline || tokens[index].Literal == ";") {
		index++
	}
	return index
}

func isTypeDeclarationKeyword(literal string) bool {
	switch literal {
	case "class", "interface", "trait", "enum", "record":
		return true
	default:
		return false
	}
}

func isModifierKeyword(literal string) bool {
	switch literal {
	case "abstract", "default", "final", "native", "private", "protected", "public", "static", "strictfp", "synchronized", "transient", "volatile":
		return true
	default:
		return false
	}
}

func isTypeKeyword(token Token) bool {
	if token.Kind != TokenKeyword {
		return false
	}
	switch token.Literal {
	case "boolean", "byte", "char", "double", "float", "int", "long", "short", "void":
		return true
	default:
		return false
	}
}

func previousIdentifierIndex(tokens []Token, start int) int {
	for i := start; i >= 0; i-- {
		if tokens[i].Kind == TokenIdentifier || isTypeKeyword(tokens[i]) {
			return i
		}
	}
	return -1
}

func modifierPrefix(tokens []Token) ([]ast.ModifierNode, []Token) {
	var modifiers []ast.ModifierNode
	index := 0
	for index < len(tokens) {
		if tokens[index].Kind == TokenKeyword && isModifierKeyword(tokens[index].Literal) {
			modifiers = append(modifiers, ast.ModifierNode{Name: tokens[index].Literal})
			index++
			continue
		}
		break
	}
	return modifiers, trimExpressionTokens(tokens[index:])
}

func declarationPrefix(tokens []Token) ([]ast.ModifierNode, []Token) {
	modifiers, rest := modifierPrefix(tokens)
	if len(rest) > 0 && rest[0].Kind == TokenKeyword && (rest[0].Literal == "def" || rest[0].Literal == "var") {
		rest = rest[1:]
	}
	return modifiers, trimExpressionTokens(rest)
}

func modifierNodes(tokens []Token) []ast.ModifierNode {
	var modifiers []ast.ModifierNode
	for _, token := range tokens {
		if token.Kind == TokenKeyword && isModifierKeyword(token.Literal) {
			modifiers = append(modifiers, ast.ModifierNode{Name: token.Literal})
		}
	}
	return modifiers
}

func hasVisibilityModifier(modifiers []ast.ModifierNode) bool {
	for _, modifier := range modifiers {
		switch modifier.Name {
		case "public", "private", "protected":
			return true
		}
	}
	return false
}

func hasDynamicDeclarationPrefix(tokens []Token) bool {
	for _, token := range tokens {
		if token.Kind == TokenKeyword && (token.Literal == "def" || token.Literal == "var") {
			return true
		}
	}
	return false
}

func typeExpression(tokens []Token) ast.Expression {
	tokens = trimExpressionTokens(tokens)
	if len(tokens) == 0 {
		return nil
	}
	return &ast.ClassExpression{Type: &ast.ConstantExpression{Kind: "type", Value: literalForTokens(tokens)}}
}

func variableDeclarationParts(tokens []Token) ([]Token, []Token) {
	tokens = trimStatementTokens(tokens)
	if len(tokens) == 0 {
		return nil, nil
	}
	if tokens[0].Kind == TokenKeyword && (tokens[0].Literal == "def" || tokens[0].Literal == "var") {
		return nil, trimExpressionTokens(tokens[1:])
	}
	boundary := len(tokens)
	for _, literal := range []string{"=", ","} {
		if idx := topLevelLiteralIndex(tokens, literal); idx >= 0 && idx < boundary {
			boundary = idx
		}
	}
	nameIndex := previousIdentifierIndex(tokens, boundary-1)
	if nameIndex <= 0 {
		return nil, nil
	}
	return trimExpressionTokens(tokens[:nameIndex]), trimExpressionTokens(tokens[nameIndex:])
}

func isImportWordToken(token Token) bool {
	return token.Kind == TokenIdentifier || token.Kind == TokenKeyword
}

func needsImportTokenSpace(prev Token, next Token) bool {
	if isImportWordToken(prev) && isImportWordToken(next) {
		return true
	}
	if prev.Literal == "*" && isImportWordToken(next) {
		return true
	}
	return false
}

func importTextForTokens(tokens []Token) string {
	b := strings.Builder{}
	tokens = trimStatementTokens(tokens)
	for i, token := range tokens {
		if i > 0 && needsImportTokenSpace(tokens[i-1], token) {
			b.WriteByte(' ')
		}
		b.WriteString(token.Literal)
	}
	return b.String()
}

func parseImport(text string) *ast.ImportNode {
	imp := &ast.ImportNode{}
	fields := strings.Fields(text)
	if len(fields) > 0 && fields[0] == "static" {
		imp.IsStatic = true
		text = strings.TrimSpace(strings.TrimPrefix(text, "static"))
	}
	if before, after, ok := strings.Cut(text, " as "); ok {
		text = strings.TrimSpace(before)
		imp.Alias = strings.TrimSpace(after)
	}
	if strings.HasSuffix(text, ".*") {
		imp.IsStar = true
		if imp.IsStatic {
			imp.ClassName = strings.TrimSuffix(text, ".*")
		} else {
			imp.PackageName = strings.TrimSuffix(text, "*")
		}
		return imp
	}
	if imp.IsStatic {
		idx := strings.LastIndex(text, ".")
		if idx >= 0 {
			imp.ClassName = text[:idx]
			imp.FieldName = text[idx+1:]
		}
		return imp
	}
	imp.ClassName = text
	return imp
}

func expressionFromToken(token Token) ast.Expression {
	switch token.Kind {
	case TokenIdentifier:
		return &ast.VariableExpression{Name: token.Literal}
	case TokenKeyword:
		switch token.Literal {
		case "true", "false":
			return &ast.ConstantExpression{Kind: "bool", Value: token.Literal}
		case "null":
			return &ast.ConstantExpression{Kind: "null", Value: token.Literal}
		case "this", "super":
			return &ast.VariableExpression{Name: token.Literal}
		default:
			return &ast.ConstantExpression{Kind: "keyword", Value: token.Literal}
		}
	case TokenNumber:
		return &ast.ConstantExpression{Kind: "number", Value: token.Literal}
	case TokenString:
		return &ast.ConstantExpression{Kind: "string", Value: token.Literal}
	default:
		return &ast.ConstantExpression{Kind: "text", Value: token.Literal}
	}
}

type parser struct {
	tokens []Token
	source string
}

func (p *parser) sourceForTokens(tokens []Token) string {
	tokens = trimExpressionTokens(tokens)
	if len(tokens) == 0 || p.source == "" {
		return literalForTokens(tokens)
	}
	start := tokens[0].Range.Start.Offset
	end := tokens[len(tokens)-1].Range.End.Offset
	if start < 0 || end < start || end > len(p.source) {
		return literalForTokens(tokens)
	}
	return strings.TrimSpace(p.source[start:end])
}

func (p *parser) parseArgumentExpression(tokens []Token) ast.Expression {
	if colon := topLevelLiteralIndex(tokens, ":"); colon > 0 {
		name := literalForTokens(tokens[:colon])
		return &ast.MapEntryExpression{Key: &ast.ConstantExpression{Kind: "string", Value: name}, Value: p.parseExpression(tokens[colon+1:])}
	}
	return p.parseExpression(tokens)
}

func (p *parser) parseArgumentListExpression(tokens []Token) *ast.ArgumentListExpression {
	args := &ast.ArgumentListExpression{}
	for _, part := range splitTopLevel(tokens, ",") {
		part = trimExpressionTokens(part)
		if len(part) == 0 {
			continue
		}
		args.Expressions = append(args.Expressions, p.parseArgumentExpression(part))
	}
	return args
}

func (p *parser) parseMethodCallExpression(tokens []Token) ast.Expression {
	if len(tokens) < 3 || tokens[0].Kind != TokenIdentifier || tokens[1].Literal != "(" {
		return nil
	}
	close := matchingClose(tokens, 1, "(", ")")
	if close < 0 || close != len(tokens)-1 {
		return nil
	}
	return &ast.MethodCallExpression{
		Object:       &ast.VariableExpression{Name: "this"},
		Method:       &ast.ConstantExpression{Kind: "string", Value: tokens[0].Literal},
		Arguments:    p.parseArgumentListExpression(tokens[2:close]),
		ImplicitThis: true,
	}
}

func (p *parser) parseMapKey(tokens []Token) ast.Expression {
	tokens = trimExpressionTokens(tokens)
	if len(tokens) == 1 && tokens[0].Kind == TokenIdentifier {
		return &ast.ConstantExpression{Kind: "string", Value: tokens[0].Literal}
	}
	return p.parseExpression(tokens)
}

func (p *parser) parseCollectionExpression(tokens []Token) ast.Expression {
	if len(tokens) < 2 || tokens[0].Literal != "[" || tokens[len(tokens)-1].Literal != "]" {
		return nil
	}
	inner := trimExpressionTokens(tokens[1 : len(tokens)-1])
	if len(inner) == 0 {
		return &ast.ListExpression{}
	}
	if topLevelLiteralIndex(inner, ":") >= 0 {
		m := &ast.MapExpression{}
		for _, part := range splitTopLevel(inner, ",") {
			part = trimExpressionTokens(part)
			if len(part) == 0 {
				continue
			}
			colon := topLevelLiteralIndex(part, ":")
			if colon < 0 {
				continue
			}
			m.Entries = append(m.Entries, ast.MapEntryExpression{Key: p.parseMapKey(part[:colon]), Value: p.parseExpression(part[colon+1:])})
		}
		return m
	}
	list := &ast.ListExpression{}
	for _, part := range splitTopLevel(inner, ",") {
		part = trimExpressionTokens(part)
		if len(part) == 0 {
			continue
		}
		list.Expressions = append(list.Expressions, p.parseExpression(part))
	}
	return list
}

func (p *parser) parseBinaryExpression(tokens []Token) ast.Expression {
	operators := []string{"=", "+=", "-=", "*=", "/=", "%=", "||", "&&", "==", "!=", "<", ">", "<=", ">=", "+", "-", "*", "/", "%"}
	for _, operator := range operators {
		idx := topLevelLiteralIndex(tokens, operator)
		if idx > 0 && idx < len(tokens)-1 {
			return &ast.BinaryExpression{Left: p.parseExpression(tokens[:idx]), Operator: operator, Right: p.parseExpression(tokens[idx+1:])}
		}
	}
	return nil
}

func (p *parser) parseExpression(tokens []Token) ast.Expression {
	tokens = trimExpressionTokens(tokens)
	if len(tokens) == 0 {
		return nil
	}
	if expr := p.parseMethodCallExpression(tokens); expr != nil {
		return expr
	}
	if expr := p.parseCollectionExpression(tokens); expr != nil {
		return expr
	}
	if expr := p.parseBinaryExpression(tokens); expr != nil {
		return expr
	}
	if len(tokens) == 1 {
		return expressionFromToken(tokens[0])
	}
	return &ast.ConstantExpression{Kind: "text", Value: p.sourceForTokens(tokens)}
}

func (p *parser) parseParameters(tokens []Token) []ast.Parameter {
	tokens = trimExpressionTokens(tokens)
	if len(tokens) == 0 {
		return nil
	}
	var parameters []ast.Parameter
	for _, part := range splitTopLevel(tokens, ",") {
		part = trimExpressionTokens(part)
		if len(part) == 0 {
			continue
		}
		_, rest := declarationPrefix(part)
		rest = trimExpressionTokens(rest)
		if len(rest) == 0 {
			continue
		}
		equal := topLevelLiteralIndex(rest, "=")
		nameSearchEnd := len(rest)
		if equal >= 0 {
			nameSearchEnd = equal
		}
		nameIndex := previousIdentifierIndex(rest, nameSearchEnd-1)
		if nameIndex < 0 {
			continue
		}
		parameter := ast.Parameter{Name: rest[nameIndex].Literal, Type: typeExpression(rest[:nameIndex])}
		if equal >= 0 {
			parameter.InitialExpression = p.parseExpression(rest[equal+1:])
		}
		parameters = append(parameters, parameter)
	}
	return parameters
}

func (p *parser) scanStatement(tokens []Token, start int) ([]Token, int) {
	depth := 0
	end := start
	for end < len(tokens) {
		token := tokens[end]
		if token.Kind == TokenEOF {
			break
		}
		if token.Kind == TokenSymbol {
			switch token.Literal {
			case "(", "[", "{":
				depth++
			case ")", "]", "}":
				if depth > 0 {
					depth--
				}
			case ";":
				if depth == 0 {
					end++
					return tokens[start:end], skipNewlines(tokens, end)
				}
			}
		}
		if token.Kind == TokenNewline && depth == 0 {
			break
		}
		end++
	}
	if end == start {
		end++
	}
	return tokens[start:end], skipNewlines(tokens, end)
}

func (p *parser) blockStatementFromTokens(tokens []Token) *ast.BlockStatement {
	open, close := enclosingPair(tokens, "{", "}")
	block := &ast.BlockStatement{}
	if open < 0 || close <= open {
		return block
	}
	inner := tokens[open+1 : close]
	for i := skipNewlines(inner, 0); i < len(inner); {
		node, next := p.parseStatement(inner, i)
		if statement, ok := node.(ast.Statement); ok {
			block.Statements = append(block.Statements, statement)
		}
		i = skipNewlines(inner, next)
	}
	return block
}

func (p *parser) parseInitializer(tokens []Token) *ast.InitializerNode {
	static := false
	start := 0
	if len(tokens) > 0 && tokens[0].Kind == TokenKeyword && tokens[0].Literal == "static" {
		static = true
		start = 1
	}
	if start < len(tokens) && tokens[start].Literal == "{" {
		return &ast.InitializerNode{Static: static, Code: p.blockStatementFromTokens(tokens[start:])}
	}
	return nil
}

func (p *parser) parseClassHeader(tokens []Token, class *ast.ClassNode) {
	for i := 0; i < len(tokens); i++ {
		if tokens[i].Kind != TokenKeyword {
			continue
		}
		switch tokens[i].Literal {
		case "extends":
			end := len(tokens)
			for j := i + 1; j < len(tokens); j++ {
				if tokens[j].Kind == TokenKeyword && tokens[j].Literal == "implements" {
					end = j
					break
				}
			}
			if class.Kind == "interface" {
				for _, part := range splitTopLevel(tokens[i+1:end], ",") {
					if expr := typeExpression(part); expr != nil {
						class.Interfaces = append(class.Interfaces, expr)
					}
				}
			} else {
				class.SuperClass = typeExpression(tokens[i+1 : end])
			}
		case "implements":
			for _, part := range splitTopLevel(tokens[i+1:], ",") {
				if expr := typeExpression(part); expr != nil {
					class.Interfaces = append(class.Interfaces, expr)
				}
			}
		}
	}
}

func (p *parser) parseCallableDeclaration(tokens []Token, className string) ast.ClassMember {
	open := topLevelOpenLiteralIndex(tokens, "(")
	if open <= 0 {
		return nil
	}
	close := matchingClose(tokens, open, "(", ")")
	if close < 0 {
		return nil
	}
	nameIndex := previousIdentifierIndex(tokens, open-1)
	if nameIndex < 0 {
		return nil
	}
	hasBody := firstNonNewlineLiteral(tokens[close+1:]) == "{"
	body := p.blockStatementFromTokens(tokens[close+1:])
	dynamicReturn := hasDynamicDeclarationPrefix(tokens[:nameIndex])
	modifiers, head := declarationPrefix(tokens[:nameIndex])
	returnTypeTokens := trimExpressionTokens(head)
	parameters := p.parseParameters(tokens[open+1 : close])
	if className != "" && len(returnTypeTokens) == 0 && tokens[nameIndex].Literal == className {
		return &ast.ConstructorNode{Modifiers: modifiers, Parameters: parameters, Code: body}
	}
	if !hasBody {
		return nil
	}
	if len(returnTypeTokens) == 0 && className != "" && !dynamicReturn {
		return nil
	}
	if len(tokens[:nameIndex]) == 0 {
		return nil
	}
	return &ast.MethodNode{
		Name:       tokens[nameIndex].Literal,
		Modifiers:  modifiers,
		ReturnType: typeExpression(returnTypeTokens),
		Parameters: parameters,
		Code:       body,
	}
}

func (p *parser) parseClassVariableMembers(tokens []Token, classKind string) []ast.ClassMember {
	modifiers, rest := modifierPrefix(tokens)
	rest = trimStatementTokens(rest)
	if len(rest) < 1 || topLevelOpenLiteralIndex(rest, "(") >= 0 {
		return nil
	}
	typeTokens, declarators := variableDeclarationParts(rest)
	if len(declarators) == 0 {
		return nil
	}
	var members []ast.ClassMember
	fieldDeclaration := classKind == "interface" || hasVisibilityModifier(modifiers)
	for _, declarator := range splitTopLevel(declarators, ",") {
		declarator = trimExpressionTokens(declarator)
		if len(declarator) == 0 {
			continue
		}
		nameIndex := firstIdentifierIndexAfter(declarator, 0)
		if nameIndex < 0 {
			continue
		}
		initialValue := ast.Expression(nil)
		if equal := topLevelLiteralIndex(declarator, "="); equal >= 0 {
			initialValue = p.parseExpression(declarator[equal+1:])
		}
		if fieldDeclaration {
			members = append(members, &ast.FieldNode{Name: declarator[nameIndex].Literal, Modifiers: modifiers, Type: typeExpression(typeTokens), InitialValueExpression: initialValue})
		} else {
			members = append(members, &ast.PropertyNode{Name: declarator[nameIndex].Literal, Modifiers: modifiers, Type: typeExpression(typeTokens), InitialValueExpression: initialValue})
		}
	}
	return members
}

func (p *parser) classMembersFromTokens(tokens []Token, className string, classKind string) []ast.ClassMember {
	tokens = trimStatementTokens(tokens)
	if len(tokens) == 0 {
		return nil
	}
	if initializer := p.parseInitializer(tokens); initializer != nil {
		return []ast.ClassMember{initializer}
	}
	for _, token := range tokens {
		if token.Kind == TokenKeyword && isTypeDeclarationKeyword(token.Literal) {
			return []ast.ClassMember{p.parseClassDeclaration(tokens)}
		}
	}
	if callable := p.parseCallableDeclaration(tokens, className); callable != nil {
		return []ast.ClassMember{callable}
	}
	return p.parseClassVariableMembers(tokens, classKind)
}

func (p *parser) parseClassMembers(tokens []Token, className string, classKind string) []ast.ClassMember {
	var members []ast.ClassMember
	for i := skipClassBodyTrivia(tokens, 0); i < len(tokens); {
		statementTokens, next := p.scanStatement(tokens, i)
		for _, member := range p.classMembersFromTokens(statementTokens, className, classKind) {
			members = append(members, member)
		}
		i = skipClassBodyTrivia(tokens, next)
	}
	return members
}

func (p *parser) parseClassDeclaration(tokens []Token) *ast.ClassNode {
	tokens = trimStatementTokens(tokens)
	keywordIndex := -1
	for i, token := range tokens {
		if token.Kind == TokenKeyword && isTypeDeclarationKeyword(token.Literal) {
			keywordIndex = i
			break
		}
	}
	if keywordIndex < 0 {
		return &ast.ClassNode{}
	}
	nameIndex := firstIdentifierIndexAfter(tokens, keywordIndex+1)
	class := &ast.ClassNode{Kind: tokens[keywordIndex].Literal, Modifiers: modifierNodes(tokens[:keywordIndex])}
	if nameIndex >= 0 {
		class.Name = tokens[nameIndex].Literal
	}
	open, close := enclosingPair(tokens, "{", "}")
	headerEnd := len(tokens)
	if open >= 0 {
		headerEnd = open
	}
	headerStart := keywordIndex + 1
	if nameIndex >= 0 {
		headerStart = nameIndex + 1
	}
	if headerStart > headerEnd {
		headerStart = headerEnd
	}
	p.parseClassHeader(tokens[headerStart:headerEnd], class)
	if open >= 0 && close > open {
		class.Members = p.parseClassMembers(tokens[open+1:close], class.Name, class.Kind)
	}
	return class
}

func (p *parser) astFromStatementTokens(tokens []Token) ast.AstNode {
	tokens = trimStatementTokens(tokens)
	if len(tokens) == 0 {
		return &ast.ExpressionStatement{}
	}
	if tokens[0].Kind == TokenKeyword {
		switch tokens[0].Literal {
		case "return":
			return &ast.ReturnStatement{Expression: p.parseExpression(tokens[1:])}
		case "throw":
			return &ast.ThrowStatement{Expression: p.parseExpression(tokens[1:])}
		case "break":
			return &ast.BreakStatement{Label: firstIdentifierAfter(tokens, 1)}
		case "continue":
			return &ast.ContinueStatement{Label: firstIdentifierAfter(tokens, 1)}
		}
	}
	for _, token := range tokens {
		if token.Kind == TokenKeyword && isTypeDeclarationKeyword(token.Literal) {
			return p.parseClassDeclaration(tokens)
		}
	}
	if method := p.parseCallableDeclaration(tokens, ""); method != nil {
		return method
	}
	return &ast.ExpressionStatement{Expression: p.parseExpression(tokens)}
}

func (p *parser) parseStatement(tokens []Token, start int) (ast.AstNode, int) {
	statementTokens, next := p.scanStatement(tokens, start)
	return p.astFromStatementTokens(statementTokens), next
}

func (p *parser) parseDirective(tokens []Token, start int, kind directiveKind) (ast.AstNode, int) {
	end := start + 1
	for end < len(tokens) && tokens[end].Kind != TokenEOF && tokens[end].Kind != TokenNewline && tokens[end].Literal != ";" {
		end++
	}
	if end < len(tokens) && tokens[end].Literal == ";" {
		end++
	}
	if kind == directivePackage {
		text := strings.TrimSpace(strings.TrimPrefix(literalForTokens(trimStatementTokens(tokens[start:end])), tokens[start].Literal))
		return &ast.PackageNode{Name: text}, skipNewlines(tokens, end)
	}
	text := importTextForTokens(tokens[start+1 : end])
	return parseImport(text), skipNewlines(tokens, end)
}

func (p *parser) significantTokens() []Token {
	result := make([]Token, 0, len(p.tokens))
	for _, token := range p.tokens {
		if token.Kind == TokenWhitespace || token.Kind == TokenComment {
			continue
		}
		result = append(result, token)
	}
	return result
}

func (p *parser) parseTopLevel(module *ast.ModuleNode) {
	tokens := p.significantTokens()
	for i := 0; i < len(tokens); {
		token := tokens[i]
		if token.Kind == TokenEOF {
			break
		}
		if token.Kind == TokenNewline {
			i++
			continue
		}
		if token.Kind == TokenKeyword && token.Literal == "package" {
			node, next := p.parseDirective(tokens, i, directivePackage)
			module.Package = node.(*ast.PackageNode)
			i = next
			continue
		}
		if token.Kind == TokenKeyword && token.Literal == "import" {
			node, next := p.parseDirective(tokens, i, directiveImport)
			module.Imports = append(module.Imports, node.(*ast.ImportNode))
			i = next
			continue
		}
		node, next := p.parseStatement(tokens, i)
		switch n := node.(type) {
		case *ast.ClassNode:
			module.Classes = append(module.Classes, n)
		case *ast.MethodNode:
			module.Methods = append(module.Methods, n)
		case ast.Statement:
			module.StatementBlock.Statements = append(module.StatementBlock.Statements, n)
		}
		i = next
	}
}

func (p *parser) validateDelimiters() error {
	type frame struct {
		literal string
		token   Token
	}
	var stack []frame
	matching := map[string]string{
		")": "(",
		"]": "[",
		"}": "{",
	}
	for _, token := range p.tokens {
		if token.Kind != TokenSymbol {
			continue
		}
		switch token.Literal {
		case "(", "[", "{":
			stack = append(stack, frame{literal: token.Literal, token: token})
		case ")", "]", "}":
			if len(stack) == 0 || stack[len(stack)-1].literal != matching[token.Literal] {
				return fmt.Errorf("unexpected %q at %d:%d", token.Literal, token.Range.Start.Line, token.Range.Start.Column)
			}
			stack = stack[:len(stack)-1]
		}
	}
	if len(stack) > 0 {
		token := stack[len(stack)-1].token
		return fmt.Errorf("unclosed %q at %d:%d", token.Literal, token.Range.Start.Line, token.Range.Start.Column)
	}
	return nil
}

func (p *parser) parse() (*ast.ModuleNode, error) {
	if err := p.validateDelimiters(); err != nil {
		return nil, err
	}
	module := &ast.ModuleNode{StatementBlock: &ast.BlockStatement{}}
	p.parseTopLevel(module)
	return module, nil
}

// Parse parses Groovy source into a Module AST.
func Parse(source string) (*ast.ModuleNode, error) {
	tokens, err := lex(source)
	if err != nil {
		return nil, err
	}
	parser := &parser{tokens: tokens, source: source}
	return parser.parse()
}
