"""Rewrite per-patch commit metadata (author, committer, times) via a callback."""

import collections.abc
import json
import os

import seed.devprod.openfork.git as git

Callable = collections.abc.Callable
Awaitable = collections.abc.Awaitable


async def update_metadata(
    updater: (
        Callable[[str, dict[str, str]], Awaitable[dict[str, str] | None]] | None
    ) = None,
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    if not updater:
        raise ValueError("updater is required")
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not os.path.isdir(open_worktree):
        raise ValueError(f"open_worktree is not a directory: {open_worktree}")

    with open(os.path.join(open_worktree, "metadata.json"), "r") as f:
        metadata = json.load(f)

    rebuild_metadata = {}
    if os.path.isfile(os.path.join(open_worktree, "rebuild.json")):
        with open(os.path.join(open_worktree, "rebuild.json"), "r") as f:
            rebuild_metadata = json.load(f)

    for patch_name in metadata:
        patch_path = os.path.join(open_worktree, f"{patch_name}.patch")
        if not os.path.isfile(patch_path):
            continue
        commit_metadata = {
            **metadata[patch_name],
            **rebuild_metadata.get(patch_name, {}),
        }
        updates = await updater(patch_name, commit_metadata)
        if updates:
            rebuild_metadata[patch_name] = {
                **rebuild_metadata.get(patch_name, {}),
                **updates,
            }
    with open(os.path.join(open_worktree, f"rebuild.json"), "w") as f:
        json.dump(rebuild_metadata, f, indent=4)

    await git.commit(open_worktree, message="update metadata\n\n" + updater.__name__)

    del kwargs
