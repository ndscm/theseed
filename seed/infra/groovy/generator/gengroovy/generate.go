package gengroovy

import (
	"fmt"
	"strings"

	"github.com/ndscm/theseed/seed/infra/groovy/ast"
)

func writeIndent(b *strings.Builder, depth int) {
	for range depth {
		b.WriteByte('\t')
	}
}

func lastRune(s string) rune {
	var result rune
	for _, r := range s {
		result = r
	}
	return result
}

func writeSeparatedNode(b *strings.Builder, node ast.AstNode, depth int) error {
	part := strings.Builder{}
	err := writeNode(&part, node, depth)
	if err != nil {
		return err
	}
	if part.Len() == 0 {
		return nil
	}
	if b.Len() > 0 && lastRune(b.String()) != '\n' {
		b.WriteByte('\n')
	}
	b.WriteString(part.String())
	return nil
}

func hasModuleBody(module *ast.ModuleNode) bool {
	return len(module.Classes) > 0 || len(module.Methods) > 0 || module.StatementBlock != nil && len(module.StatementBlock.Statements) > 0
}

func writeModule(b *strings.Builder, module *ast.ModuleNode, depth int) error {
	if module.Package != nil {
		err := writeNode(b, module.Package, depth)
		if err != nil {
			return err
		}
		b.WriteByte('\n')
		if len(module.Imports) > 0 || hasModuleBody(module) {
			b.WriteByte('\n')
		}
	}

	for _, imp := range module.Imports {
		err := writeNode(b, imp, depth)
		if err != nil {
			return err
		}
		b.WriteByte('\n')
	}
	if len(module.Imports) > 0 && hasModuleBody(module) {
		b.WriteByte('\n')
	}

	for _, class := range module.Classes {
		err := writeSeparatedNode(b, class, depth)
		if err != nil {
			return err
		}
	}
	for _, method := range module.Methods {
		err := writeSeparatedNode(b, method, depth)
		if err != nil {
			return err
		}
	}
	if module.StatementBlock != nil {
		for _, statement := range module.StatementBlock.Statements {
			err := writeSeparatedNode(b, statement, depth)
			if err != nil {
				return err
			}
		}
	}
	return nil
}

func writeImport(b *strings.Builder, imp *ast.ImportNode) {
	b.WriteString("import ")
	if imp.IsStatic {
		b.WriteString("static ")
	}
	if imp.IsStatic {
		b.WriteString(imp.ClassName)
		if imp.IsStar {
			b.WriteString(".*")
		} else if imp.FieldName != "" {
			b.WriteByte('.')
			b.WriteString(imp.FieldName)
		}
	} else if imp.IsStar {
		b.WriteString(imp.PackageName)
		b.WriteByte('*')
	} else {
		b.WriteString(imp.ClassName)
	}
	if imp.Alias != "" {
		b.WriteString(" as ")
		b.WriteString(imp.Alias)
	}
}

func writeExpressionList(b *strings.Builder, expressions []ast.Expression, depth int) error {
	for i, expression := range expressions {
		if i > 0 {
			b.WriteString(", ")
		}
		err := writeNode(b, expression, depth)
		if err != nil {
			return err
		}
	}
	return nil
}

func writeModifiers(b *strings.Builder, modifiers []ast.ModifierNode) {
	for _, modifier := range modifiers {
		b.WriteString(modifier.Name)
		b.WriteByte(' ')
	}
}

func writeParameters(b *strings.Builder, parameters []ast.Parameter) {
	for i, parameter := range parameters {
		if i > 0 {
			b.WriteString(", ")
		}
		if parameter.Type != nil {
			_ = writeNode(b, parameter.Type, 0)
			b.WriteByte(' ')
		}
		b.WriteString(parameter.Name)
		if parameter.InitialExpression != nil {
			b.WriteString(" = ")
			_ = writeNode(b, parameter.InitialExpression, 0)
		}
	}
}

func writeTypeOrDef(b *strings.Builder, typ ast.Expression, depth int) {
	if typ == nil {
		b.WriteString("def")
		return
	}
	_ = writeNode(b, typ, depth)
}

func writeConstructor(b *strings.Builder, constructor *ast.ConstructorNode, name string, depth int) error {
	writeModifiers(b, constructor.Modifiers)
	b.WriteString(name)
	b.WriteByte('(')
	writeParameters(b, constructor.Parameters)
	b.WriteString(") ")
	return writeNode(b, constructor.Code, depth)
}

func writeClassMember(b *strings.Builder, class *ast.ClassNode, member ast.ClassMember, depth int) error {
	if constructor, ok := member.(*ast.ConstructorNode); ok {
		return writeConstructor(b, constructor, class.Name, depth)
	}
	return writeNode(b, member, depth)
}

func writeClass(b *strings.Builder, class *ast.ClassNode, depth int) error {
	b.WriteString(class.Kind)
	b.WriteByte(' ')
	b.WriteString(class.Name)
	if class.SuperClass != nil {
		b.WriteString(" extends ")
		if err := writeNode(b, class.SuperClass, depth); err != nil {
			return err
		}
	}
	if len(class.Interfaces) > 0 {
		if class.Kind == "interface" {
			b.WriteString(" extends ")
		} else {
			b.WriteString(" implements ")
		}
		if err := writeExpressionList(b, class.Interfaces, depth); err != nil {
			return err
		}
	}
	b.WriteString(" {")
	if len(class.Members) == 0 {
		b.WriteByte('}')
		return nil
	}
	b.WriteByte('\n')
	for _, member := range class.Members {
		writeIndent(b, depth+1)
		err := writeClassMember(b, class, member, depth+1)
		if err != nil {
			return err
		}
		b.WriteByte('\n')
	}
	writeIndent(b, depth)
	b.WriteByte('}')
	return nil
}

func writeMethod(b *strings.Builder, method *ast.MethodNode, depth int) error {
	writeModifiers(b, method.Modifiers)
	writeTypeOrDef(b, method.ReturnType, depth)
	b.WriteByte(' ')
	b.WriteString(method.Name)
	b.WriteByte('(')
	writeParameters(b, method.Parameters)
	b.WriteString(") ")
	return writeNode(b, method.Code, depth)
}

func writeField(b *strings.Builder, field *ast.FieldNode, depth int) error {
	writeModifiers(b, field.Modifiers)
	writeTypeOrDef(b, field.Type, depth)
	b.WriteByte(' ')
	b.WriteString(field.Name)
	if field.InitialValueExpression != nil {
		b.WriteString(" = ")
		return writeNode(b, field.InitialValueExpression, depth)
	}
	return nil
}

func writeProperty(b *strings.Builder, property *ast.PropertyNode, depth int) error {
	writeModifiers(b, property.Modifiers)
	writeTypeOrDef(b, property.Type, depth)
	b.WriteByte(' ')
	b.WriteString(property.Name)
	if property.InitialValueExpression != nil {
		b.WriteString(" = ")
		return writeNode(b, property.InitialValueExpression, depth)
	}
	return nil
}

func writeInitializer(b *strings.Builder, initializer *ast.InitializerNode, depth int) error {
	if initializer.Static {
		b.WriteString("static ")
	}
	return writeNode(b, initializer.Code, depth)
}

func writeBlockStatement(b *strings.Builder, block *ast.BlockStatement, depth int) error {
	b.WriteByte('{')
	if len(block.Statements) == 0 {
		b.WriteByte('}')
		return nil
	}
	b.WriteByte('\n')
	for _, statement := range block.Statements {
		writeIndent(b, depth+1)
		err := writeNode(b, statement, depth+1)
		if err != nil {
			return err
		}
		b.WriteByte('\n')
	}
	writeIndent(b, depth)
	b.WriteByte('}')
	return nil
}

func isWhitespaceByte(ch byte) bool {
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

func isCompoundEquals(text string, index int) bool {
	if index < 0 || index >= len(text) || text[index] != '=' {
		return false
	}
	if index > 0 && strings.ContainsRune("=!<>+-*/%", rune(text[index-1])) {
		return true
	}
	return index+1 < len(text) && text[index+1] == '='
}

func isTextWordByte(ch byte) bool {
	return ('a' <= ch && ch <= 'z') || ('A' <= ch && ch <= 'Z') || ('0' <= ch && ch <= '9') || ch == '_' || ch == '$' || ch == '\'' || ch == '"'
}

func needsTextTokenSpace(text string, index int, previous byte, next byte) bool {
	if isWhitespaceByte(next) {
		return false
	}
	if previous == ',' && next != ')' && next != ']' && next != '}' {
		return true
	}
	if previous == ':' && next != ')' && next != ']' && next != '}' {
		return true
	}
	if next == '=' {
		return !isCompoundEquals(text, index)
	}
	if previous == '=' {
		return !isCompoundEquals(text, index-1)
	}
	if (previous == ')' || previous == ']' || previous == '}') && isTextWordByte(next) {
		return true
	}
	return false
}

func quotedTextEnd(text string, start int) int {
	quote := text[start]
	triple := start+2 < len(text) && text[start+1] == quote && text[start+2] == quote
	if triple {
		for i := start + 3; i < len(text); i++ {
			if text[i] == '\\' {
				i++
				continue
			}
			if i+2 < len(text) && text[i] == quote && text[i+1] == quote && text[i+2] == quote {
				return i + 2
			}
		}
		return len(text) - 1
	}
	for i := start + 1; i < len(text); i++ {
		if text[i] == '\\' {
			i++
			continue
		}
		if text[i] == quote {
			return i
		}
	}
	return len(text) - 1
}

func endsWithWhitespace(b *strings.Builder) bool {
	if b.Len() == 0 {
		return false
	}
	ch := b.String()[b.Len()-1]
	return ch == ' ' || ch == '\t' || ch == '\r' || ch == '\n'
}

func formatTextExpression(text string) string {
	b := strings.Builder{}
	var previous byte
	hasPrevious := false
	for i := 0; i < len(text); i++ {
		ch := text[i]
		if isWhitespaceByte(ch) {
			b.WriteByte(ch)
			if ch == '\n' {
				hasPrevious = false
			}
			continue
		}
		if ch == '\'' || ch == '"' {
			if hasPrevious && needsTextTokenSpace(text, i, previous, ch) && !endsWithWhitespace(&b) {
				b.WriteByte(' ')
			}
			end := quotedTextEnd(text, i)
			b.WriteString(text[i : end+1])
			previous = text[end]
			hasPrevious = true
			i = end
			continue
		}
		if hasPrevious && needsTextTokenSpace(text, i, previous, ch) && !endsWithWhitespace(&b) {
			b.WriteByte(' ')
		}
		b.WriteByte(ch)
		previous = ch
		hasPrevious = true
	}
	return b.String()
}

func endsWithNewline(b *strings.Builder) bool {
	if b.Len() == 0 {
		return false
	}
	return b.String()[b.Len()-1] == '\n'
}

func writeFormattedNewline(b *strings.Builder, atLineStart *bool, pendingSpace *bool) {
	trimTrailingHorizontalWhitespace(b)
	if !endsWithNewline(b) {
		b.WriteByte('\n')
	}
	*atLineStart = true
	*pendingSpace = false
}

func writeFormattedIndent(b *strings.Builder, depth int, atLineStart *bool) {
	if !*atLineStart {
		return
	}
	writeIndent(b, depth)
	*atLineStart = false
}

func lastByte(b *strings.Builder) byte {
	if b.Len() == 0 {
		return 0
	}
	return b.String()[b.Len()-1]
}

func writePendingSpace(b *strings.Builder, pendingSpace *bool, next byte) {
	if !*pendingSpace {
		return
	}
	previous := lastByte(b)
	if previous != 0 && previous != '\n' && previous != '\t' && previous != ' ' && previous != '(' && previous != '[' && next != ',' && next != ')' && next != ']' {
		b.WriteByte(' ')
	}
	*pendingSpace = false
}

func writeFormattedQuotedText(b *strings.Builder, text string, index int, atLineStart *bool, pendingSpace *bool) int {
	writeFormattedIndent(b, 0, atLineStart)
	writePendingSpace(b, pendingSpace, text[index])
	end := quotedTextEnd(text, index)
	b.WriteString(text[index : end+1])
	return end
}

func writeFormattedLineComment(b *strings.Builder, text string, index int, depth int, atLineStart *bool, pendingSpace *bool) int {
	writeFormattedIndent(b, depth, atLineStart)
	writePendingSpace(b, pendingSpace, '/')
	end := index
	for end < len(text) && text[end] != '\n' && text[end] != '\r' {
		end++
	}
	b.WriteString(text[index:end])
	return end - 1
}

func endsWithLineIndent(b *strings.Builder) bool {
	text := b.String()
	for i := len(text) - 1; i >= 0; i-- {
		switch text[i] {
		case '\n':
			return true
		case '\t':
			continue
		default:
			return false
		}
	}
	return true
}

func needsSpaceBeforeOpenBrace(b *strings.Builder) bool {
	if b.Len() == 0 || endsWithLineIndent(b) {
		return false
	}
	text := b.String()
	last := text[len(text)-1]
	return last != ' ' && last != '\n' && last != '\t'
}

func trimTrailingHorizontalWhitespace(b *strings.Builder) {
	text := b.String()
	end := len(text)
	for end > 0 && (text[end-1] == ' ' || text[end-1] == '\t') {
		end--
	}
	if end == len(text) {
		return
	}
	b.Reset()
	b.WriteString(text[:end])
}

func nextNonWhitespaceByte(text string, start int) byte {
	for i := start; i < len(text); i++ {
		if !isWhitespaceByte(text[i]) {
			return text[i]
		}
	}
	return 0
}

func ensureTrailingCommaBeforeClosingList(b *strings.Builder) {
	text := b.String()
	if text == "" || text[len(text)-1] != '\n' {
		return
	}
	end := len(text) - 1
	for end > 0 && (text[end-1] == ' ' || text[end-1] == '\t') {
		end--
	}
	if end == 0 {
		return
	}
	last := text[end-1]
	if last == '[' || last == ',' {
		return
	}
	b.Reset()
	b.WriteString(text[:end])
	b.WriteByte(',')
	b.WriteByte('\n')
}

func writeBracedText(b *strings.Builder, text string, depth int) {
	currentDepth := depth
	parenDepth := 0
	bracketDepth := 0
	bracketIndents := []bool{}
	preclosedBrackets := 0
	atLineStart := true
	pendingSpace := false
	lineDepth := func() int {
		return currentDepth + parenDepth + bracketDepth
	}
	popBracketIndent := func() bool {
		if len(bracketIndents) == 0 {
			return false
		}
		last := len(bracketIndents) - 1
		contributed := bracketIndents[last]
		if contributed && bracketDepth > 0 {
			bracketDepth--
		}
		bracketIndents = bracketIndents[:last]
		return contributed
	}
	precloseLineBrackets := func(index int) int {
		closers := 0
		contributed := false
		for next := index; next < len(text) && text[next] == ']'; next++ {
			contributed = popBracketIndent() || contributed
			preclosedBrackets++
			closers++
		}
		if closers == 1 && !contributed && bracketDepth > 0 {
			return 1
		}
		return 0
	}
	for i := 0; i < len(text); i++ {
		switch text[i] {
		case '\'', '"':
			i = writeFormattedQuotedText(b, text, i, &atLineStart, &pendingSpace)
		case '/':
			if i+1 < len(text) && text[i+1] == '/' {
				i = writeFormattedLineComment(b, text, i, lineDepth(), &atLineStart, &pendingSpace)
				continue
			}
			writeFormattedIndent(b, lineDepth(), &atLineStart)
			writePendingSpace(b, &pendingSpace, text[i])
			b.WriteByte(text[i])
		case '{':
			writeFormattedIndent(b, lineDepth(), &atLineStart)
			trimTrailingHorizontalWhitespace(b)
			if needsSpaceBeforeOpenBrace(b) {
				b.WriteByte(' ')
			}
			b.WriteByte('{')
			currentDepth++
			writeFormattedNewline(b, &atLineStart, &pendingSpace)
			if i+1 < len(text) && text[i+1] == '\n' {
				i++
			}
		case '}':
			currentDepth--
			if currentDepth < depth {
				currentDepth = depth
			}
			if !atLineStart {
				writeFormattedNewline(b, &atLineStart, &pendingSpace)
			}
			writeFormattedIndent(b, lineDepth(), &atLineStart)
			b.WriteByte('}')
			if next := nextNonWhitespaceByte(text, i+1); next != 0 && next != '}' && next != ')' && next != ']' && next != ',' {
				writeFormattedNewline(b, &atLineStart, &pendingSpace)
			}
		case '(':
			writeFormattedIndent(b, lineDepth(), &atLineStart)
			writePendingSpace(b, &pendingSpace, text[i])
			b.WriteByte('(')
			parenDepth++
		case ')':
			if parenDepth > 0 {
				parenDepth--
			}
			writeFormattedIndent(b, lineDepth(), &atLineStart)
			writePendingSpace(b, &pendingSpace, text[i])
			b.WriteByte(')')
		case '[':
			writeFormattedIndent(b, lineDepth(), &atLineStart)
			writePendingSpace(b, &pendingSpace, text[i])
			contributesIndent := lastByte(b) != '[' && bracketDepth == 0
			b.WriteByte('[')
			if contributesIndent {
				bracketDepth++
			}
			bracketIndents = append(bracketIndents, contributesIndent)
		case ']':
			closeDepthAdjustment := 0
			if atLineStart {
				ensureTrailingCommaBeforeClosingList(b)
				closeDepthAdjustment = precloseLineBrackets(i)
			}
			if preclosedBrackets > 0 {
				preclosedBrackets--
			} else {
				popBracketIndent()
			}
			writeFormattedIndent(b, lineDepth()-closeDepthAdjustment, &atLineStart)
			writePendingSpace(b, &pendingSpace, text[i])
			b.WriteByte(']')
		case ',':
			writeFormattedIndent(b, lineDepth(), &atLineStart)
			trimTrailingHorizontalWhitespace(b)
			b.WriteByte(',')
		case '\n':
			if !atLineStart {
				writeFormattedNewline(b, &atLineStart, &pendingSpace)
			}
		case '\r', '\t', ' ':
			if !atLineStart {
				pendingSpace = true
			}
		default:
			writeFormattedIndent(b, lineDepth(), &atLineStart)
			writePendingSpace(b, &pendingSpace, text[i])
			b.WriteByte(text[i])
		}
	}
	trimTrailingHorizontalWhitespace(b)
}

func writeConstantExpression(b *strings.Builder, expr *ast.ConstantExpression, depth int) error {
	text := expr.Value
	if expr.Kind == "text" {
		text = formatTextExpression(text)
	}
	if expr.Kind == "text" && strings.ContainsAny(expr.Value, "{}") {
		writeBracedText(b, text, depth)
		return nil
	}
	b.WriteString(text)
	return nil
}

func writeDeclarationExpression(b *strings.Builder, expr *ast.DeclarationExpression, depth int) error {
	if err := writeNode(b, expr.Left, depth); err != nil {
		return err
	}
	b.WriteByte(' ')
	b.WriteString(expr.Operator)
	b.WriteByte(' ')
	return writeNode(b, expr.Right, depth)
}

func expressionNeedsMultiline(expression ast.Expression) bool {
	switch expr := expression.(type) {
	case *ast.ConstantExpression:
		return strings.Contains(expr.Value, "\n")
	case *ast.MapEntryExpression:
		return expressionNeedsMultiline(expr.Key) || expressionNeedsMultiline(expr.Value)
	case *ast.ArgumentListExpression:
		return expressionsNeedMultiline(expr.Expressions)
	case *ast.ListExpression:
		return expressionsNeedMultiline(expr.Expressions)
	case *ast.MapExpression:
		for _, entry := range expr.Entries {
			if expressionNeedsMultiline(entry.Key) || expressionNeedsMultiline(entry.Value) {
				return true
			}
		}
	}
	return false
}

func expressionsNeedMultiline(expressions []ast.Expression) bool {
	for _, expression := range expressions {
		if expressionNeedsMultiline(expression) {
			return true
		}
	}
	return false
}

func writeMultilineArgumentListExpression(b *strings.Builder, args *ast.ArgumentListExpression, depth int) error {
	b.WriteByte('\n')
	for i, expression := range args.Expressions {
		writeIndent(b, depth+1)
		if err := writeNode(b, expression, depth+1); err != nil {
			return err
		}
		if i < len(args.Expressions)-1 {
			b.WriteByte(',')
		}
		b.WriteByte('\n')
	}
	writeIndent(b, depth)
	return nil
}

func writeArgumentListExpression(b *strings.Builder, args *ast.ArgumentListExpression, depth int) error {
	if expressionsNeedMultiline(args.Expressions) {
		return writeMultilineArgumentListExpression(b, args, depth)
	}
	return writeExpressionList(b, args.Expressions, depth)
}

func writeListExpression(b *strings.Builder, list *ast.ListExpression, depth int) error {
	b.WriteByte('[')
	err := writeExpressionList(b, list.Expressions, depth)
	if err != nil {
		return err
	}
	b.WriteByte(']')
	return nil
}

func writeMapExpression(b *strings.Builder, m *ast.MapExpression, depth int) error {
	b.WriteByte('[')
	for i, entry := range m.Entries {
		if i > 0 {
			b.WriteString(", ")
		}
		err := writeNode(b, entry.Key, depth)
		if err != nil {
			return err
		}
		b.WriteString(": ")
		err = writeNode(b, entry.Value, depth)
		if err != nil {
			return err
		}
	}
	b.WriteByte(']')
	return nil
}

func writeMethodCallExpression(b *strings.Builder, call *ast.MethodCallExpression, depth int) error {
	if !call.ImplicitThis && call.Object != nil {
		err := writeNode(b, call.Object, depth)
		if err != nil {
			return err
		}
		b.WriteByte('.')
	}
	err := writeNode(b, call.Method, depth)
	if err != nil {
		return err
	}
	b.WriteByte('(')
	if call.Arguments != nil {
		err = writeNode(b, call.Arguments, depth)
		if err != nil {
			return err
		}
	}
	b.WriteByte(')')
	return nil
}

func writeBinaryExpression(b *strings.Builder, expr *ast.BinaryExpression, depth int) error {
	err := writeNode(b, expr.Left, depth)
	if err != nil {
		return err
	}
	b.WriteByte(' ')
	b.WriteString(expr.Operator)
	b.WriteByte(' ')
	return writeNode(b, expr.Right, depth)
}

func writeNode(b *strings.Builder, node ast.AstNode, depth int) error {
	switch n := node.(type) {
	case nil:
		return nil
	case *ast.ModuleNode:
		return writeModule(b, n, depth)
	case *ast.PackageNode:
		b.WriteString("package ")
		b.WriteString(n.Name)
	case *ast.ImportNode:
		writeImport(b, n)
	case *ast.ClassNode:
		return writeClass(b, n, depth)
	case *ast.MethodNode:
		return writeMethod(b, n, depth)
	case *ast.ConstructorNode:
		return writeConstructor(b, n, "this", depth)
	case *ast.FieldNode:
		return writeField(b, n, depth)
	case *ast.PropertyNode:
		return writeProperty(b, n, depth)
	case *ast.InitializerNode:
		return writeInitializer(b, n, depth)
	case *ast.BlockStatement:
		return writeBlockStatement(b, n, depth)
	case *ast.ExpressionStatement:
		return writeNode(b, n.Expression, depth)
	case *ast.ReturnStatement:
		b.WriteString("return")
		if n.Expression != nil {
			b.WriteByte(' ')
			return writeNode(b, n.Expression, depth)
		}
	case *ast.ThrowStatement:
		b.WriteString("throw ")
		return writeNode(b, n.Expression, depth)
	case *ast.BreakStatement:
		b.WriteString("break")
		if n.Label != "" {
			b.WriteByte(' ')
			b.WriteString(n.Label)
		}
	case *ast.ContinueStatement:
		b.WriteString("continue")
		if n.Label != "" {
			b.WriteByte(' ')
			b.WriteString(n.Label)
		}
	case *ast.VariableExpression:
		b.WriteString(n.Name)
	case *ast.ConstantExpression:
		return writeConstantExpression(b, n, depth)
	case *ast.ClassExpression:
		return writeNode(b, n.Type, depth)
	case *ast.DeclarationExpression:
		return writeDeclarationExpression(b, n, depth)
	case *ast.ArgumentListExpression:
		return writeArgumentListExpression(b, n, depth)
	case *ast.MapEntryExpression:
		if err := writeNode(b, n.Key, depth); err != nil {
			return err
		}
		b.WriteString(": ")
		return writeNode(b, n.Value, depth)
	case *ast.ListExpression:
		return writeListExpression(b, n, depth)
	case *ast.MapExpression:
		return writeMapExpression(b, n, depth)
	case *ast.MethodCallExpression:
		return writeMethodCallExpression(b, n, depth)
	case *ast.BinaryExpression:
		return writeBinaryExpression(b, n, depth)
	default:
		return fmt.Errorf("groovy: unknown node type %T", node)
	}
	return nil
}

func formatIndent(source string, indent string) string {
	if source == "" {
		return ""
	}
	lines := strings.Split(source, "\n")
	for i, line := range lines {
		depth := 0
		for depth < len(line) && line[depth] == '\t' {
			depth++
		}
		lines[i] = strings.Repeat(indent, depth) + line[depth:]
	}
	return strings.Join(lines, "\n")
}

// GenerateIndent renders Groovy source from a parsed AST with indentation.
func GenerateIndent(node ast.AstNode, indent string) (string, error) {
	b := strings.Builder{}
	err := writeNode(&b, node, 0)
	if err != nil {
		return "", err
	}
	return formatIndent(b.String(), indent), nil
}

// Generate renders Groovy source from a parsed AST using two spaces for indentation.
func Generate(node ast.AstNode) (string, error) {
	return GenerateIndent(node, "  ")
}
