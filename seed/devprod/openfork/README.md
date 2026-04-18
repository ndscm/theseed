# openfork

Toolset for manipulating a series of `format-patch` files — filtering, editing,
and rewriting commits at the patch level instead of through interactive rebase.

> One use case is producing an open-source fork of a monorepo.

## How it works

The pipeline has three phases — export, manipulate, and rebuild — with a staging
worktree in between that keeps every patch edit under version control.

- `export`: dump each commit on the source branch as a numbered `.patch` file in
  a staging worktree (`/tmp/open`), and record the original author and committer
  info in `metadata.json`.
- `manipulate`: each step edits those patches in place and commits the result,
  so the edits themselves live in git history.
- `rebuild`: apply the edited patches one by one to produce a fresh history in a
  separate worktree (`/tmp/rebuild`).

## Modules

- [export.py](export.py) — `format-patch` the main branch, capture metadata.
- [prepare.py](prepare.py) — create the project branch in the staging worktree.
- [diffpath.py](diffpath.py) — keep or drop whole file diffs by path regex.
- [diffedit.py](diffedit.py) — rewrite hunk lines in files by content regex.
- [walk.py](walk.py) — generic per-patch visitor for custom rewrites.
- [scalpel.py](scalpel.py) — replace or hand-edit a single named patch.
- [message.py](message.py) — rewrite commit messages.
- [metadata.py](metadata.py) — rewrite per-commit author/committer metadata.
- [clean.py](clean.py) — drop patches that became empty after edits.
- [rebuild.py](rebuild.py) — replay patches into the final output worktree.
- [patch.py](patch.py) — unified-diff parser used by the editors.
- [git.py](git.py) — thin `subprocess` wrappers for git operations.
- [worktree.py](worktree.py) — resolves and creates the worktrees.
