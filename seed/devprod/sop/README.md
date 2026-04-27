# SOP

**SOP** (Standard Operating Procedures) is the canonical home for the coding
standards, conventions, and tooling guidelines that every coding factory worker
is expected to follow when working in the theseed monorepo and in any repository
downstream of it.

"Downstream" here means any repository that consumes theseed code, vendors parts
of it, or is generated from it. The same SOPs apply in those repositories so
that code produced anywhere in the theseed ecosystem stays consistent — a coding
factory worker should not need to switch styles when crossing a repository
boundary.

The documents here serve coding factory workers in two modes at once:

- **As a reference during code review and onboarding.** When a coding factory
  worker asks "why is this laid out this way?", the answer should live in
  `sop/`.
- **As prompt for coding.** The same documents are loaded into the working
  context of any coding factory worker that produces code, so that anything
  affecting how generated code should look, organize itself, or use shared
  tooling is applied consistently across workflows.

Serving both modes shapes how SOP documents are written: they are
self-contained, declarative, and prescriptive rather than discursive.

## Scope

SOP covers project-wide engineering conventions. Topics that fit here:

- Source-level layout rules (file organization, ordering, naming).
- Language idioms that the project deliberately standardizes on.
- Standard tools and how they should be invoked.
- Shared patterns for cross-cutting concerns (errors, logging, configuration).

Topics that do **not** belong here:

- Per-package design notes — keep those in the package's own `README.md`.
- Architecture decisions for a single subsystem — keep those next to the
  subsystem.
- Runbooks for production incidents — those live with the operational tooling
  for the affected service.

If a rule only applies to one package or one repository, it isn't an SOP; it's a
local convention and belongs next to the code it governs.

## How to Use SOP

1. **Before writing new code**, skim the documents that touch the area you are
   changing. The short examples in each doc are enough to internalize the
   pattern.
2. **During code review**, cite the relevant SOP by filename when requesting a
   change. This makes the standard auditable rather than personal preference.

## Authoring New SOP Documents

A new document is justified when a convention applies broadly across the theseed
monorepo and its downstream repositories and is not already enforced by a linter
or formatter. (If a tool can enforce it automatically, prefer the tool.)

Each SOP file follows the same shape, mirroring the existing documents:

1. **Title (H1)** — name the convention, not the file. For example,
   `# Function Ordering: Strict Proximity Bottom-Up`.
2. **Abstract** — one to three sentences stating the rule and the reasoning
   behind it. A reader should be able to act on this paragraph alone.
3. **Rules** — numbered or bulleted, written in the imperative. State what is
   required, not what is suggested.
4. **Code Example** — one realistic snippet that illustrates the rule. Prefer
   examples drawn from the languages the rule applies to (Go for Go-specific
   rules, etc.).
5. **Pros & Cons** — be honest about the trade-offs. The existing documents call
   out drawbacks explicitly; new ones should too.

### Style Conventions for SOP Documents

- Write in present-tense, declarative prose.
- Keep examples minimal. They illustrate the rule, not a full program.
- Prefer concrete file names, function names, and types over abstractions like
  `Foo` and `Bar` — concrete examples are easier to internalize.
- Wrap prose at roughly 80 columns to match the rest of the repository's
  Markdown.
