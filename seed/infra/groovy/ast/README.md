# Groovy AST

This package defines the Groovy AST model used by our formatter. The node names
and package split follow the official Groovy compiler AST as closely as
practical:

https://github.com/apache/groovy/tree/master/src/main/java/org/codehaus/groovy/ast

## Formatter-Specific Differences

This AST is intentionally smaller than the compiler AST. It only models syntax
that the formatter needs to parse, preserve, and print. Compiler-only state such
as source units, redirect metadata, resolved Java classes, type-checking caches,
and transformation metadata is omitted.

The main structural difference is class member ordering. Groovy's `ClassNode`
stores members in separate collections, such as fields, properties,
constructors, and methods. Those collections preserve order within each member
kind, but they do not preserve the original mixed order of the class body.

For formatting, mixed order matters. A source class may interleave fields,
methods, properties, and initializers, and the formatter must not reorder them.
To preserve that source order, our `ClassNode` includes a mixed `Members` layer
in addition to the official-style member node types.
