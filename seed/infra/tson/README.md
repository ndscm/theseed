# tson

tson (TypeScript Object Notation) is a configuration format that uses TypeScript
files (`.config.ts`) as the source of truth. The files provide full TypeScript
type-checking and editor support but are parsed statically — no code is ever
executed.

## Motivation

JSON lacks comments, trailing commas, and type safety. YAML is error-prone. TOML
doesn't scale to nested structures. TypeScript already has the best editor
tooling (autocomplete, inline validation, go-to-definition), so tson reuses it
for configuration while keeping the runtime properties of a static data format.

## How it works

A tson file is a valid TypeScript file whose only runtime-meaningful part is a
single `export default` expression containing a JSON-compatible object literal.
Everything else — imports, type annotations, `satisfies` clauses, and comments —
is stripped during parsing.

Parsers in any language can extract the configuration by:

1. Removing TypeScript-only syntax (type imports, type annotations, `satisfies`)
   and comments.
2. Extracting the `export default` expression.
3. Wrapping unquoted object keys in double quotes to produce valid JSON.
4. Parsing the result as JSON.

Because no code is executed, tson files are safe to consume from any language
without a JavaScript runtime.

## Example

A tson config with an imported type:

```ts
import { type User } from "./type/user"

export default {
  id: 1,
  name: "Nagi",
} satisfies User
```

The type file (`type/user.ts`) is standard TypeScript:

```ts
export type User = {
  id: number
  name: string
}
```

Types can also be defined inline:

```ts
type User = {
  id: number
  name: string
}

export default {
  id: 1,
  name: "Nagi",
} satisfies User
```

In all cases the parsed output is the same plain object:

```json
{
  "id": 1,
  "name": "Nagi"
}
```

## Rules

- The file must contain exactly one `export default` expression.
- The exported value must be a JSON-compatible object literal (strings, numbers,
  booleans, nulls, arrays, and plain objects).
- Imports are skipped during parsing — only the `export default` value matters.
  Type-only imports (`import { type Foo }` or `import type { Foo }`) are
  recommended for clarity, but the parser does not enforce this.
- No function calls, template literals, variable references, or any other
  runtime expressions.
