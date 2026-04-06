"""Delete patches that have no remaining diffs."""

import json
import logging
import os

import seed.devprod.openfork.git as git
import seed.devprod.openfork.patch as patch

logger = logging.getLogger(__name__)


async def clean_empty(*, open_worktree: str = "", **kwargs) -> None:
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not os.path.isdir(open_worktree):
        raise ValueError(f"open_worktree is not a directory: {open_worktree}")

    with open(os.path.join(open_worktree, "metadata.json"), "r") as f:
        metadata = json.load(f)

    removed = []
    for patch_name in list(metadata):
        patch_path = os.path.join(open_worktree, f"{patch_name}.patch")
        if not os.path.isfile(patch_path):
            continue

        with open(patch_path, "r", encoding="utf-8") as f:
            text = f.read()

        p = patch.Patch(os.path.basename(patch_path), text)
        if not p.diffs:
            logger.info("Removing empty patch %s", os.path.basename(patch_path))
            os.remove(patch_path)
            removed.append(patch_name)

    await git.commit(
        open_worktree,
        message="clean empty\n\n" + "\n".join(removed),
        skip_empty=True,
    )

    del kwargs
