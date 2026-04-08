"""Iterate every patch and apply a user callback to edit contents and metadata."""

import collections.abc
import json
import logging
import os

import seed.devprod.openfork.git as git
import seed.devprod.openfork.patch as patch

Callable = collections.abc.Callable
Awaitable = collections.abc.Awaitable

logger = logging.getLogger(__name__)


async def walk(
    updater: Callable[
        [str, patch.Patch, dict],
        Awaitable[tuple[patch.Patch | None, dict | None]],
    ],
    *,
    open_worktree: str = "",
    open_project: str = "",
    **kwargs,
) -> None:
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not open_project:
        raise ValueError("open_project is required")

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

        with open(patch_path, "r", encoding="utf-8") as f:
            original_text = f.read()

        p = patch.Patch(os.path.basename(patch_path), original_text)
        new_p, updates = await updater(
            patch_name,
            p,
            {
                **metadata[patch_name],
                **rebuild_metadata.get(patch_name, {}),
            },
        )
        if new_p:
            result = new_p.render()
            if result != original_text:
                with open(patch_path, "w", encoding="utf-8") as f:
                    f.write(result)
        if updates:
            rebuild_metadata[patch_name] = {
                **rebuild_metadata.get(patch_name, {}),
                **updates,
            }
    with open(os.path.join(open_worktree, f"rebuild.json"), "w") as f:
        json.dump(rebuild_metadata, f, indent=4)

    await git.commit(open_worktree, message="walk\n\n" + updater.__name__)

    del kwargs
