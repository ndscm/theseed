# Type Organization: Strict Contiguity

This document defines the structural conventions for organizing custom types,
their constructors, and their methods within a Go package.

## 1. Strict Contiguity of Types and Methods

In Go, it is easy for methods belonging to the same `type` to become scattered
throughout a file or spread across multiple files. To maintain strict
encapsulation and ensure readability, a `type` definition and **all** of its
associated methods must be kept tightly grouped in a single, uninterrupted
block.

### The Contiguity Rules:

1. **Single-File Scope:** A type definition and all of its receiver methods must
   reside in the exact same file. You must never define a type in one file and
   implement some of its methods in another.
2. **No Interleaving:** Absolutely no other type definitions, package-level
   variables, or standalone functions may appear between the type definition and
   its methods, or between the methods themselves. The block must be contiguous.
3. **Layout Order:**
   - **First:** The `struct` or `interface` declaration.
   - **Second:** The `NewXXX` constructor / factory (if applicable).
   - **Third:** All receiver methods bound to that type.

### Code Example

```go
package oidc

// ==========================================
// CORRECT: Uninterrupted Type Block
// ==========================================

// 1. Type Definition
type KeycloakAuthenticator struct {
    clientID string
    timeout  time.Duration
}

// 2. Constructor (Immediately follows type)
func NewKeycloakAuthenticator(clientID string) *KeycloakAuthenticator {
    return &KeycloakAuthenticator{
        clientID: clientID,
        timeout:  5 * time.Second,
    }
}

// 3. Receiver Methods (Tightly grouped, no interleaving)
func (k *KeycloakAuthenticator) Authenticate(token string) error {
    // ... auth logic ...
    return nil
}

func (k *KeycloakAuthenticator) refresh() error {
    // ... refresh logic ...
    return nil
}

// ==========================================
// STOP: The block for KeycloakAuthenticator is now closed.
// New types or vars can only appear after all methods are defined.
// ==========================================

var DefaultScopes = []string{"openid", "profile"}

type TokenResponse struct {
    // ...
}
```

## 2. Initialization: `NewXXX` vs. `Initialize`

When setting up a new object, you must choose exactly one of two initialization
patterns. **A type should expose either a `NewXXX` function OR an `Initialize`
method, but never both.**

### Pattern A: The `New` Function (Simple Setup)

Use a `NewXXX` function when the struct only requires basic memory allocation,
default value assignments, or simple field injection.

**Constraints for `NewXXX`:**

- **Must return `*XXX` only.** It must never return an `error`.
- **No heavy setup.** It should not perform network calls, file I/O, complex
  parsing, or block the thread.

```go
type SandboxConfig struct {
    ImagePath string
    RootPriv  bool
}

// NewSandboxConfig acts as a static factory.
// It does not return an error and performs no I/O.
func NewSandboxConfig(image string) *SandboxConfig {
    return &SandboxConfig{
        ImagePath: image,
        RootPriv:  false, // Assign simple defaults
    }
}

func (s *SandboxConfig) IsValid() bool {
    return s.ImagePath != ""
}
```

### Pattern B: The `Initialize` Method (Heavy Setup)

If setting up the object requires reaching out to a database, parsing
configuration files, or resolving network addresses, the initialization might
fail. In these cases, omit the `NewXXX` function entirely. Instead, expose an
`Initialize` method that the caller invokes after creating the struct literal
themselves.

**Constraints for `Initialize`:**

- **Must return an `error`.**
- **Used for side-effects and heavy lifting.**
- **Replaces the need for a `NewXXX` factory.**

```go
type DatabaseClient struct {
    DSN  string
    conn *sql.DB
}

// Note: There is NO NewDatabaseClient function here.
// The caller creates the struct literal and calls Initialize().

// Initialize handles the heavy lifting that might fail.
func (d *DatabaseClient) Initialize(ctx context.Context) error {
    if d.DSN == "" {
        return errors.New("DSN cannot be empty")
    }

    // Simulating heavy setup (e.g., dialing a connection)
    db, err := sql.Open("pgx", d.DSN)
    if err != nil {
        return fmt.Errorf("failed to connect to database: %w", err)
    }

    d.conn = db
    return nil
}

func (d *DatabaseClient) Query(query string) error {
    // ... query logic ...
    return nil
}
```

## Pros

- **Zero Fragmentation:** By enforcing single-file scope and preventing
  interleaving, developers never have to hunt across a package to find out what
  a type can or cannot do. The visual bounding box mimics standard
  object-oriented encapsulation.
- **Clear Error Contracts:** By strictly forbidding `NewXXX` from returning
  errors, callers instantly know whether instantiating a type is a "safe"
  operation (simple memory allocation) or a "risky" operation requiring an
  `Initialize` call with proper error handling.
- **Forces Single Responsibility:** If a type's contiguous block becomes too
  massive to read comfortably in one file, it is an architectural signal that
  the type is doing too much and should be broken down into smaller, composed
  types.

## Cons

- **Long Type Blocks:** A type with many methods produces a single long,
  uninterruptible block. There is no escape valve like splitting methods into
  `foo_query.go` / `foo_mutate.go` — the block has to live as one. _(Mitigation:
  treat the size pressure as the architectural signal called out above and
  decompose the type instead of the file.)_
- **No Mixed Initialization Path:** A type must commit to either `NewXXX` or
  `Initialize` for its entire lifetime. If most callers only need a cheap
  constructor but a small subset later needs heavy setup, neither option fits
  cleanly and you typically have to introduce a second type. The rule prevents
  ambiguity at the cost of some flexibility.
- **Exposed Pre-Init State for `Initialize` Types:** Because callers build the
  struct literal themselves before calling `Initialize`, the zero-value form of
  the struct is publicly observable. Methods on `Initialize` types must defend
  against being called on an uninitialized receiver, or the type must keep
  setup-required fields unexported and document the call order.
- **Friction with Generated or Build-Tagged Code:** Tooling that emits methods
  into a separate file (mocks, code generators, build-tag-specific
  implementations) violates the single-file rule by default. These cases need
  explicit carve-outs in the generator configuration or a different
  organizational boundary (e.g., a sibling type that wraps the generated one).
