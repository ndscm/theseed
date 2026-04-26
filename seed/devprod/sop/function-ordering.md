# Function Ordering: Strict Proximity Bottom-Up

When organizing functions within a single file, this project follows the
**Strict Proximity Bottom-Up** ordering pattern (also known as Depth-First,
Post-Order Traversal).

The core philosophy is **Spatial Locality**: code that runs together should live
together, and a reader should never encounter a function call before they have
seen its definition.

## The Core Rules

1. **Bottom-Up Root:** The primary public API or entry point of the file is
   placed at the **very bottom**.
2. **Define Before Use:** A function is never called before it is defined.
3. **Strict Proximity:** Helper functions are placed _immediately above_ the
   specific function that consumes them.
4. **Depth-First Grouping:** Do not group all low-level helpers at the top of
   the file. Instead, group them by the logical branch of the dependency tree
   they serve.

### Handling Shared Helpers

If a helper function is used by _multiple_ distinct branches within the same
file, place it above the **first common ancestor** of the functions that use it,
or at the very top of the file if it is a truly global utility.

## Code Example

Below is an example of how this layout creates distinct, self-contained blocks
of logic:

```go
package service

// ==========================================
// BLOCK 1: Data Retrieval Branch
// ==========================================

func buildUserQuery(id string) string {
    return "SELECT * FROM users WHERE id = " + id
}

func executeDBQuery(query string) (User, error) {
    // ... database execution ...
}

// fetchUser relies strictly on the functions immediately above it.
func fetchUser(userID string) (User, error) {
    query := buildUserQuery(userID)
    return executeDBQuery(query)
}

// ==========================================
// BLOCK 2: Validation Branch
// ==========================================

// validateRequest has no relationship to fetchUser,
// so it is placed here, serving only the root function below.
func validateRequest(req Request) error {
    // ... validation logic ...
}

// ==========================================
// BLOCK 3: Business Logic Branch
// ==========================================

func executeTransfer(user User, amount float64) error {
    // ... business logic ...
}

// ==========================================
// ROOT: The Public API
// ==========================================

// ProcessTransaction is the ultimate root.
// The functions directly above it are exactly the ones that serve it.
func ProcessTransaction(req Request) error {
    if err := validateRequest(req); err != nil {
        return err
    }

    user, err := fetchUser(req.UserID)
    if err != nil {
        return err
    }

    return executeTransfer(user, req.Amount)
}
```

## Pros

- **Self-Contained Refactoring:** Because a function and its exclusive
  dependencies are physically grouped, extracting logic into a new package or
  file is trivial. You can highlight a continuous block of code, cut, and paste
  without leaving orphaned helpers behind.
- **Mental Garbage Collection:** As developers read top-to-bottom, they can
  safely drop leaf-node functions from their working memory as soon as they
  finish reading the parent function.
- **Zero Forward Referencing:** You never experience the cognitive interruption
  of seeing a function call and having to scroll around to figure out what it
  does. The context is established before the execution.
- **No False Relationships:** It prevents the mixing of unrelated helper
  functions. Physical adjacency guarantees a logical dependency.

## Cons

- **Entry Point is Hidden:** The biggest drawback is that the "headline" of the
  file (the public API) is at the bottom. New contributors opening the file for
  the first time will see low-level implementation details first and must scroll
  to the bottom to find the entry point. _(Mitigation: Modern IDE symbol
  outlines and "Go to Definition" features make this less of an issue in
  practice)._
- **Friction with Shared Helpers:** Strict proximity is easy when a helper has
  one parent. If a helper is used in five different places throughout the file,
  deciding exactly where to place it requires slightly more thought to avoid
  breaking the flow.
- **Learning Curve:** It runs counter to the popular "Newspaper Metaphor"
  (Top-Down) ordering that many developers are taught, requiring a slight
  paradigm shift during onboarding and code reviews.
