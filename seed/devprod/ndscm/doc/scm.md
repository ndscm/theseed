# SCM: The Source Code Management System

An SCM manages a shared directory of mostly plain text — typically source code.
Popular implementations include git, svn, and hg. An SCM exists to solve two
fundamental problems:

1. Record the change history of a shared directory.
2. Reconcile conflicts when independent changes to that directory must be
   merged.

## The concept of the history of a shared directory

Every legendary project begins with an empty directory. We call this the
_history root_.

From there, the directory grows step by step. Most people picture each step as a
snapshot of the directory's state at that point in time. That picture is only
half right.

A step is more precisely a _delta_ between two snapshots, and a delta can be
expressed in patch format.

A patch hunk is a tuple of:

- the base snapshot it applies to,
- the file within the shared directory it modifies,
- the line range it targets,
- the lines it adds,
- the lines it removes.

A patch file is a collection of patch hunks that share the same base snapshot.

When patches chain cleanly from the empty root forward, every workstation can
replay the chain to reconstruct the same snapshot deterministically.

To change the shared directory, a contributor creates a patch and shares it,
either with themselves later or with others, to be applied.

In strict mode, a patch must be applied to its declared base snapshot. In
practice, people want to apply patches onto _different_ base snapshots. Because
a patch is immutable, this requires an algorithm and tooling that generate a
_new_ patch from the original patch and the desired base snapshot.

The most widely used tool for this is `git rebase`. It uses the surrounding
context lines of each hunk as hints to locate the corresponding region in the
new base. When the algorithm cannot determine the location automatically, it
hands control back to the user to author the new hunks manually — what we call
_resolving conflicts_.

When people speak of "the same patch" before and after a rebase, that is
technically incorrect: the base snapshot has changed, the line numbers have
changed, and only the added/removed content tends to survive intact. The rebased
patch is a new patch.

Each SCM is, at its core, a piece of software that defines:

- how patches are stored, in a database or on the filesystem;
- how users create and share patches.

Because patches are immutable, "updating a patch" does not exist — only creating
a new patch, either from scratch or from an existing patch plus a new base.

## Common workflow

A snapshot identified by a content hash is hard to remember, so people attach
named pointers to snapshots. In git terminology:

- a _branch_ is a named pointer that may be moved to point at a different
  snapshot later;
- a _tag_ is a named pointer that is fixed once set.

Each project owner typically claims write permission over a single
source-of-truth named pointer and asks contributors to base their patches on the
snapshot that pointer currently identifies. The dominant convention, inherited
from git, is to call this pointer `main`.

The following walks through how this plays out with git and GitHub.

A contributor fetches every patch in the project's history and reconstructs the
latest snapshot in their worktree:

```bash
git clone https://github.com/org/repo.git
```

They create a personal named pointer to author patches against the
source-of-truth pointer, then write patches from scratch:

```bash
git checkout -b name/dev origin/main   # create a personal named pointer for authoring patches against the source-of-truth named pointer
# make changes
git commit -a -m "patch 1"             # create the first patch from scratch
# make changes
git commit -a -m "patch 2"             # create the second patch from scratch
```

They share the patches by publishing the pointer and ask the owner to advance
the source-of-truth pointer to their snapshot:

```bash
git push origin name/dev
```

If the owner accepts the contribution, they usually do not point `main` directly
at the contributor's snapshot. Instead, they recreate the patches under the
owner's committer identity:

```bash
# Performed on the owner's side, typically by GitHub on merge.
git fetch origin                       # fetch the patches shared by the contributor
git checkout main                      # prepare to advance the "main" named pointer
git cherry-pick main..name/dev         # recreate the patches based on the main snapshot and the contributor's patches
git push origin main                   # advance the owner-assigned source-of-truth named pointer
```

The owner may instead ask the contributor to rebase onto the latest `main`
first. In that case the contributor runs:

```bash
git fetch origin                       # fetch the latest snapshot that the source-of-truth named pointer points at
git checkout name/dev                  # load the original patches
git rebase origin/main                 # recreate the patches based on the new base snapshot and the original patches
git push origin -f name/dev            # share the recreated patches under the updated personal named pointer
```

## Conclusion

The patch is the primitive of source code management. A snapshot is what you get
by replaying a chain of patches from the history root, and a named pointer is a
label attached to a snapshot. Every production SCM — git, svn, hg, fossil,
perforce — is a specific set of engineering choices about how patches are stored
and how patches are shared between contributors.

With this model in hand, the everyday operations of any particular tool stop
being arbitrary commands to memorize and reveal themselves as instances of two
underlying actions: creating a patch, and regenerating a patch on top of a
different base snapshot.
