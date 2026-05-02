# Error Discovery: Use Seed Error and Log Libraries

Native language errors are optimized for performance: they are compact and
intentionally omit details like stack traces and source locations. This makes
production debugging difficult. Additionally, error conventions differ
substantially across languages (Go's `error` interface, TypeScript's `Error`
class, etc.), creating inconsistency in multi-language monorepos.

The theseed monorepo provides two libraries that solve both problems:

- **`seed/infra/error/go/seederr`** — wraps errors with a full goroutine stack
  trace and an optional numeric error code.
- **`seed/infra/log/go/seedlog`** — logs messages with the exact source file and
  line number of the call site.

These libraries minimize the gap between languages by enforcing a consistent
error shape: every error carries a stack, and every log line carries a source
location.

## The Rules

1. **Always wrap errors with `seederr`.** Use `seederr.Wrap(err)` when
   propagating an existing error. Use `seederr.WrapErrorf(...)` when creating a
   new error with a message. Never return a bare `fmt.Errorf` or `errors.New` —
   the stack trace will be missing.

2. **Always log with `seedlog`.** Use `seedlog.Infof`, `seedlog.Warnf`,
   `seedlog.Errorf`, and `seedlog.Debugf` for all application logging. Never use
   `log.Printf` or `fmt.Printf` — the source location will be missing.

3. **Attach an error code to expose clear error semantics.** (Optional) Use
   `seederr.Code(code, err)` or `seederr.CodeErrorf(code, ...)` when you want
   calling code to programmatically distinguish error cases. This is not
   mandatory — the decision is up to the developer based on whether the caller
   benefits from a machine-readable code. The code can be:
   - A **globally unique application error code** — these propagate end-to-end
     and can be tested against in any language (e.g. a webapp asserting a
     specific numeric code returned by the backend).
   - A standard **Abseil/gRPC status code** (e.g. 5 for `NOT_FOUND`) —
     acceptable for convenience when a unique code is not needed.

   The `seedgrpc` log interceptor extracts the low byte as the gRPC status code
   and strips internal details before sending the response to the client.

   Use `seederr.DefaultCode(code, err)` when propagating an error that may or
   may not already carry a code. It only fills in the code if one has not been
   set, so it will not overwrite a more specific code attached deeper in the
   call stack.

4. **Never expose internal error details to callers.** The `seedgrpc`
   interceptor logs the full `SeedError` (stack trace, message) server-side,
   then returns only the unwrapped error message and gRPC code to the client.
   Service code should rely on this boundary — do not manually construct
   `connect.NewError` with verbose messages.

## Code Example

```go
package service

import (
    "github.com/ndscm/theseed/seed/infra/error/go/seederr"
    "github.com/ndscm/theseed/seed/infra/log/go/seedlog"
)

func fetchDocument(id string) (*Document, error) {
    row, err := db.Query(id)
    if err != nil {
        return nil, seederr.Wrap(err)
    }
    if row == nil {
        // Code 5 = NOT_FOUND in gRPC. The interceptor maps this to the
        // client-facing gRPC status; the message stays server-side.
        return nil, seederr.CodeErrorf(5, "document %s not found", id)
    }
    seedlog.Infof("Fetched document: %s", id)
    return row, nil
}
```

When `fetchDocument` returns an error, the `SeedError` carries:

- The original error message.
- A goroutine stack trace captured at the wrap site.
- An optional gRPC code (if attached).

The `seedgrpc` log interceptor prints the full error server-side, then sends
only `connect.NewError(NOT_FOUND, <unwrapped error>)` to the client.

## Pros

- **Uniform Diagnostics:** Every error in every service carries a stack trace.
  Debugging a production issue never requires reproducing the call path
  manually.
- **Source-Located Logging:** Log lines include the file and line number of the
  call site, eliminating "which log statement produced this?" ambiguity.
- **Language Consistency:** The same error shape (message + stack + code)
  applies across Go packages in the monorepo. Adding new languages follows the
  same contract.
- **Automatic Boundary Enforcement:** The gRPC interceptor guarantees that
  internal details (stack traces, verbose messages) never leak to external
  callers without requiring discipline from every service author.

## Cons

- **Stack Capture Cost:** `debug.Stack()` is called on every `Wrap`. For
  extremely hot paths that produce many errors per second, this has measurable
  overhead. _(Mitigation: `seederr.Wrap` short-circuits if the error is already
  a `SeedError`, so re-wrapping at each call layer is free.)_
- **Verbose Local Output:** In development, the full stack trace in every error
  can produce noisy terminal output. _(Mitigation: the ANSI-dimmed formatting in
  `SeedError.Error()` visually de-emphasizes the stack relative to the
  message.)_
- **Code Discipline:** Developers must remember to use `seederr` and `seedlog`
  instead of stdlib equivalents. A bare `fmt.Errorf` compiles without complaint
  but silently drops the stack. _(Mitigation: code review against this SOP, and
  grep-based lint checks for bare error construction in service code.)_
