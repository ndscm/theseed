# TypeScript Buildinfo

Unlike other languages, build info for TypeScript is not embedded directly into
the package. TypeScript packages are typically bundled in a single Vite build,
and embedding volatile build info would break hermeticity and cause cache
misses. Instead, the hermetic boundary is pushed to runtime: the HTTP server
injects build info into the HTML page, and this package reads it from there.
