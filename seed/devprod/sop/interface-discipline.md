# Interface Discipline: Avoid Overusing Interfaces

An interface decouples a caller from an implementation. It takes many forms:
language-native interfaces, generics, inheritance, and subclassing within a
process, and dynamic-link libraries, IPC, HTTP APIs, and RPC across a process
boundary. Whatever the form, an interface buys exactly two things: the freedom
to swap in different implementations (plugins) and the ability to hide
implementation details from the caller. Neither benefit is free — both cost
readability, because the reader can no longer jump from a call to the code that
runs. In the ndscm ecosystem, readability is the fundamental requirement, so an
interface is justified only when one of those two benefits is actually needed.

## Background

Of the two benefits, some schools of thought treat the first — swappable
implementations — as an unconditional good, and dependency injection is the
pattern built on that premise: wire code against interfaces so any concrete type
can be substituted from the outside. Java SPI is the purest expression of it,
deriving nearly every class from an interface so that any of them _could_ be
replaced later.

Applied by default, this is a disaster for the fundamental requirement of the
ndscm ecosystem: the readability of code. The implementation is hidden behind an
interface that, in practice, has exactly one implementation. A reader can no
longer jump from a call to the code that runs; they have to search the
repository for the real type instead of letting the language server take them
there directly. Every reader pays that navigation cost for a flexibility no one
actually uses. Decoupling is a benefit only when the flexibility is real.

## Rules

1. **Introduce an interface only for a real reason.** Add one when the system
   genuinely needs multiple interchangeable implementations (a plugin system) or
   deliberately hides the implementation from the call site (a stable API across
   a process boundary). Absent one of these, call the concrete type directly.
2. **Do not decouple speculatively.** Reject the dependency-injection reflex of
   deriving a class from an interface "just in case" a second implementation
   appears later. Add the interface when the second implementation actually
   arrives, not in case it might.
3. **Keep the call path navigable.** A reader with a language server should
   reach the running code with a single "go to definition." Every interface
   between a caller and its sole implementation turns that jump into a
   repository-wide search, so pay that cost only when a rule above earns it.
4. **A cross-process boundary is not an automatic license.** Dynamic-link
   libraries, IPC, HTTP, and RPC all present interfaces, but the boundary alone
   does not justify layering extra in-process interfaces behind it. Apply the
   same test to the code on each side.

## Code Example

```go
package scm

type ScmProvider interface {
    Clone(ctx context.Context, url string) error
}

type GitProvider struct{}

func (g *GitProvider) Clone(ctx context.Context, url string) error {
    return nil
}

var _ ScmProvider = (*GitProvider)(nil)

type SvnProvider struct{}

func (g *SvnProvider) Clone(ctx context.Context, url string) error {
    return nil
}

var _ ScmProvider = (*SvnProvider)(nil)
```

## Pros

- **Direct navigation.** Without a needless interface in the path, a reader
  jumps from any call straight to the code that executes, keeping the language
  server's "go to definition" reliable.
- **Honest architecture.** The presence of an interface signals a real extension
  point. Readers can trust that an interface means "expect multiple
  implementations" rather than "someone added a layer by reflex."
- **Less indirection to hold in mind.** Fewer hops between caller and
  implementation means less scaffolding to trace before understanding what the
  code does.

## Cons

- **Retrofitting cost.** When a second implementation genuinely appears later,
  the concrete type must be lifted into an interface and its callers updated —
  work that a speculative interface would have paid for up front.
- **Judgment required.** "Real reason" is not something a linter can enforce, so
  the call falls to author and reviewer, and reasonable people will sometimes
  disagree about whether a boundary qualifies.
- **Testing friction.** Depending on a concrete type can make substituting a
  fake in tests harder than an interface would. Weigh this honestly, but do not
  treat "it's easier to mock" as automatic grounds for an interface when the
  seam exists only for the test.
