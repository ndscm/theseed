# TypeScript in Theseed

TypeScript is supported for both frontend and Node.js. Raw JavaScript is
generally discouraged across the theseed ecosystem. Compiler options in
`tsconfig.json` are generated with `tsc --init --extraOptions` to keep a
consistent style and format throughout the monorepo.

`paths` is generally discouraged because `tsc` does not rewrite them into
relative imports in its output. While `esbuild` plugins can handle this, we
chose not to pull in extra dependencies for a single style concern. Relative
imports are required across the monorepo for best compatibility. This rule does
not apply to `tson` config files (`*.ndscm.ts`).
