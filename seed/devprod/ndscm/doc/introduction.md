# ndscm: Introduction

Git has been the dominant SCM for years, but it supports so many workflows that
it's hard to learn. While experienced engineers can master git, most users treat
it as little more than a manual save-and-snapshot tool. Building on the concepts
introduced in SCM Concept (scm.md), we propose an opinionated workflow for
contributing to a team repository. This workflow cuts through the complexity of
managing countless commands and flags. We surface only the opinionated workflow
through the ndscm CLI, hiding the underlying SCM system behind it.

## Core Concept

### Built around patches

The entire ndscm workflow is built around the concept of a "patch". Although a
patch is implemented as a git commit, a patch in the ndscm ecosystem carries
additional metadata. Every patch has a change-uuid, a stable identifier that
remains unchanged even when the patch is reapplied to a different snapshot (e.g.
during a rebase). This stability gives ndscm ecosystem tools the knowledge they
need to control behavior around each patch. The patch metadata is stored in git
commit message trailers. The standard metadata fields are:

- **Change-uuid**: A UUID that identifies the patch. Generated when a patch is
  first crafted. It stays the same even when the patch is reapplied to a
  different snapshot.
- **Break**: A message indicating that this patch breaks existing behavior. Must
  appear together with "Migrate".
- **Migrate**: A message describing how to resolve the breaking change. Must
  appear together with "Break".
- **Side-effect-of-change-uuid**: Indicates that this patch exists because a
  migration is required on the branch. Must point to a change-uuid whose patch
  carries both "Break" and "Migrate", and must appear immediately after that
  breaking change. A chain of side-effect patches only triggers tests at the
  final snapshot.
- **Test-for-change-uuid**: Indicates that this patch is a unit, integration, or
  e2e test for the core behavior change. Note that existing tests broken by the
  change should use "Side-effect-of-change-uuid", not this field.
- **Change-branch**: The repository URL and branch that this change targets. The
  merge queue verifies this field before advancing the branch pointer.
- **Change-review**: The URL of the code review request. Written by the merge
  queue; may appear multiple times.
- **Change-list**: The URL of the applied code after review. Written by the
  merge queue; may appear multiple times.
- **Upstream-change**: The upstream repository URL and change identifier (e.g.
  commit hash for an upstream git repo). Present when this change is a
  minimally-edited reapplication of an upstream change. Reviewers should focus
  on security concerns and breaking changes, not on how the code is written.
- **Skip-test**: The reason this patch should not trigger tests. Omit this field
  when the reason is already conveyed by "Side-effect-of-change-uuid".

With this metadata, we move away from the traditional pattern of grouping
patches (e.g. bundling tests or side effects into the same commit as the core
change). Breaking a large commit into much smaller, metadata-annotated patches
makes each one small enough for a reviewer to understand on its own, without
burying the core logic under a pile of side effects. It also gives finer control
over reviewer assignment across a large monorepo: each reviewer can focus
exclusively on changes within their owned module.

Each code change request typically contains a chain of small patches. The merge
queue applies the patches one by one and advances the branch pointer at once,
without squashing the submitted patches. It's the contributor's responsibility
to split a large, hard-to-review patch into small patches where each one does
exactly one thing. Don't mix side effects and tests into the core change. The
contributor is also responsible for writing the metadata of each patch.

Within the ndscm ecosystem, all patches must keep a linear history. No patch may
have two or more parent snapshots. Merge commits are not allowed.

When two change chains need to converge, there are two directions. Both produce
a linear history — no traditional merge commits are created in either case.

- **Rebase mode**: rebase our patches onto the head of their chain. Our patches
  are regenerated; their chain stays unchanged. This is the normal way a
  contributor updates a feature branch against the latest shared branch.
- **Melt mode**: rebase their patches onto the head of our chain. Their patches
  are regenerated; our chain stays unchanged. This is how a shared branch
  absorbs upstream changes without rewriting its own history. The incoming
  changes go through the same test and code review process as normal
  contributions, with only minimal edits to satisfy vendor compliance
  requirements.

## Terms

### (Patch) Hunk

A patch hunk is a tuple of:

- the base snapshot it applies to,
- the file within the shared directory it modifies,
- the line range it targets,
- the lines it adds,
- the lines it removes.

Saved in patch hunk format.

### Patch

A patch is a list of patch hunks applied to the same base snapshot.

Saved in patch file format.

### Single File Patch

A single file patch is a patch that modifies or renames exactly one file.

### Change

A change is a patch with metadata. The standard metadata fields are:

- Change-title
- Change-message
- Change-uuid
- Break
- Migrate
- Side-effect-of-change-uuid
- Test-for-change-uuid
- Change-branch
- Change-review
- Change-list
- Upstream-change
- Skip-test

When using ndscm with git, the change-title is stored as the commit title, the
change-message is stored as the commit body (excluding trailers), and all other
metadata fields are stored as git commit trailers.

Since a change and a patch have a one-to-one correspondence, the terms "change"
and "patch" may be used interchangeably throughout the rest of this document.

### Change Chain

A change chain is an immutable snapshot of the shared directory. It grows from
an empty directory by applying one change after another.

### Change Chain Pointer (Branch/Tag)

A change chain pointer is a short identifier for a change chain snapshot. The
pointer is shared across the team within the SCM system. If the pointer can be
modified and the modification can be shared, it is called a "branch". If the
pointer is immutable, it is called a "tag".

### Change List (CL)

A change list is a list of changes that are applied to the shared branch pointer
all together — the pointer updates once, regardless of how many changes the list
contains. This rule does not apply to reverting: contributors or maintainers may
revert individual changes from a change list without reverting the entire list.
See the best practice section on constructing a change list.

### Change Request (CR)

A change request (or code review) is a request to change the shared branch
pointer, where the old pointer is an ancestor of the new pointer (fast forward).
A change request must have a single change list attached, along with reviewer
notifications, review comments, and discussion around the change list. Note that
the change list is immutable: updating a change request means attaching a new
change list to the request, not modifying the existing one.

### Rebase Request (RR)

A rebase request (or rebase review) is a request to change a shared branch
pointer where the new pointer is not an ancestor of the old pointer. In addition
to the information in a normal change request, a rebase request must specify the
new merge base snapshot. Two change lists are attached: the first covers
upstream changes between the old merge base and the new merge base; the second
covers the rebased changes between the old merge base and the old branch head.
The first change list is only required when the rebase request moves forward
(not backward).

## The Ordinary Development Workflow

To start contributing to a project, whether ndscm-managed or not, first clone
the repository to local disk:

```bash
nd connect git@github.com:org/repo.git
```

Rather than creating a repository workspace directly, `nd` creates a repo home
directory and uses worktrees for different branches. From there, the contributor
can create a development branch:

```bash
cd $ND_REPOS_HOME/repo/main/
nd dev # creates a repo/dev/ worktree and automatically cd to it
```

All development should happen on top of the dev branch in this worktree, grouped
by change list:

```bash
# develop a core patch
nd commit -m "<title>" --break "<break-prompt>" --migrate "<migrate-prompt>"
# develop a side effect
nd commit -m "<title>" --side-effect-of HEAD^
# develop a test
nd commit -m "<title>" --test-for HEAD^^
nd cut "<change-request-name>" HEAD # cut the change list for code review later
# develop the next issue
nd commit
nd commit ...
nd cut "..." HEAD
```

Select and submit a change list for code review when the contributor feels
ready:

```bash
nd submit "<change-request-name>"
```

Note that `nd submit` doesn't push the branch to the review server directly. It
picks the changes in the change list, reapplies them against the latest main
branch, and pushes the resulting change list to the review server to create the
change request.

When the contributor needs to modify committed changes — whether on their own
initiative or at a reviewer's request — the following commands are available:

```bash
# drop: drop the change from the change list
nd drop "<change>"

# amend: stop after the target change with a clean worktree
# new worktree changes will be folded in as a fixup when `nd continue`
nd amend "<change>"

# split: stop before the target change with target changes applied as dirty files in worktree
# new worktree changes will become a new commit before the target change when `nd continue`
nd split "<change>"

# shift: move the position of a change in the change list
# the shift target is a tuple of direction, count (default: 0), flags (default: none), e.g.
# - <2 move towards the root commit 2 times
# - >3! move towards the dev branch head 3 times, move out of the current change list into the next change list
# - >! move to the first change of the next change list
nd shift "<change>" "<target>"
```

After the change list is carved perfectly, the contributor can submit the change
list again.

## Best Practice

### Constructing a change list

A well-constructed change list is self-describing across its changes. The
patches should appear in the following order:

1. Preparatory changes for the core change (e.g. adjusting visibility, fixing
   bugs that affect the core change)
2. Behavior change 1
3. Side-effect changes for behavior change 1 (only if behavior change 1 is a
   breaking change)
4. Test changes for behavior change 1
5. Additional core changes, each followed by its side-effect (if breaking) and
   test changes
6. The terminal behavior change whose effect motivates the entire change list
7. Its side-effect (if breaking) and test changes
8. Cleanup changes

Each change should cover only one project, so that the project reviewer only
needs to examine the changes within their scope, using the rest of the changes
as reference context.

## The Motivation behind ndscm

Many developers hit a wall when working on a large project with a large team.
The traditional workflow is to create a change branch from the main branch, make
changes, push for review, wait for review — anywhere from seconds to weeks — and
then create another change branch on top of the updated main branch. This
workflow has several pain points:

- The review process becomes a blocking step in feature development.
- Developing one change on top of another pending change isn't feasible.
- Contributors frequently run into conflicts between their own branches when a
  pending review updates the implementation before it's merged.
- These self-conflicts are hard to resolve when pending changes land on the main
  branch in unexpected orders.
- Throwing a large code change at reviewers drastically slows down the review
  process.
- Developers can't see the full picture of a feature until the dependent code
  lands, so they're unable to refine their code before submitting for review,
  leaving the pending code less mature than it could be.

Given these problems, and since code review is a hard requirement for many
teams, the solution we propose is:

- Decouple the development process from the review process.
- Develop whenever you have time; submit whichever part you want for review
  whenever you're ready.

To achieve this, we arrived at a few core design principles:

- Patches must be as small as possible.
- Reapplying a patch should be painless (avoid conflicts as much as possible).
- The tooling should be able to recognize the same change and not bother the
  developer with that kind of conflict, surfacing only real conflicts.
- When a conflict does reach the developer, the resolution should be as obvious
  as possible.

With these ideas in mind, the ndscm tool was born.

The initial version of ndscm was a hobby project built with a bunch of bash
scripts. The author presented a short slide as a SETI at the "Google Internal
Devops Summit 2018". The idea was iced and not advertised after the author
became an exGoogler and the CTO of small start-ups, where code review was a
rubber stamp that took only seconds. The idea resurfaced when we started the
"LLM team coding" project (known as "Harness" today) in 2024, as we wanted
multiple LLM models to review each other's code. ndscm became the core tool for
regulating how both human and agent contributors make contributions on the team.
