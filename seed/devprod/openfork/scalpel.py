"""Edit or replace a single named patch in place."""

import collections.abc
import json
import os

import seed.devprod.openfork.git as git
import seed.devprod.openfork.patch as patch

Callable = collections.abc.Callable
Awaitable = collections.abc.Awaitable


async def replace_patch(
    patch_name: str,
    patch_text: str,
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    if not patch_name:
        raise ValueError("patch_name is required")
    if not patch_text:
        raise ValueError("patch_text is required")
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not os.path.isdir(open_worktree):
        raise ValueError(f"open_worktree is not a directory: {open_worktree}")

    with open(os.path.join(open_worktree, f"{patch_name}.patch"), "w") as f:
        f.write(patch_text)

    await git.commit(open_worktree, message="replace patch:" + patch_name)

    del kwargs


async def edit_patch(
    patch_name: str,
    updater: Callable[
        [str, patch.Patch, dict],
        Awaitable[tuple[patch.Patch | None, dict | None]],
    ],
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    if not patch_name:
        raise ValueError("patch_name is required")
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

    patch_path = os.path.join(open_worktree, f"{patch_name}.patch")
    with open(patch_path, "r", encoding="utf-8") as f:
        original_text = f.read()

    p = patch.Patch(patch_name, original_text)
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
        with open(os.path.join(open_worktree, "rebuild.json"), "w") as f:
            json.dump(rebuild_metadata, f, indent=4)

    await git.commit(open_worktree, message="replace patch:" + patch_name)

    del kwargs
