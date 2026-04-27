# Groovy

This package provides a lightweight Groovy parser and source generator for the
parts of Groovy used in this repository, including Jenkins scripted and
declarative pipeline files.

The parser reads Groovy source into the local AST model in `ast/`. The generator
in `generator/gengroovy/` renders that AST back to Groovy source. Parsing and
then generating the same file acts like a formatter: it normalizes indentation,
spacing, block layout, and common Jenkins pipeline list formatting while
preserving source constructs that are not fully modeled yet as text nodes.

## Components

- `lexer.go`, `token.go`, `keyword.go`, and `symbol.go`: tokenize Groovy source.
- `parser.go`: builds a `*ast.ModuleNode` from tokens.
- `ast/`: defines the AST nodes used by parser and generator.
- `generator/gengroovy/`: renders AST nodes back to Groovy source.

## Scope

This is not intended to be a complete Groovy compiler front end. It focuses on
repository needs: imports, package declarations, classes, methods, fields,
common expressions, collection literals, method calls, and Jenkins pipeline
blocks. Unsupported or ambiguous expressions may be retained as text so they can
round-trip through the formatter path.
