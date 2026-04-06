"""Pick or remove file diffs in every patch by regex on their a/b paths."""

import json
import logging
import os
import re

import seed.devprod.openfork.git as git
import seed.devprod.openfork.patch as patch

logger = logging.getLogger(__name__)


async def _pick_rename(
    a_patterns: list[re.Pattern[str]],
    b_patterns: list[re.Pattern[str]],
    *,
    open_worktree: str = "",
    open_project: str = "",
    open_main_branch: str = "",
    **kwargs,
) -> None:
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not os.path.isdir(open_worktree):
        raise ValueError(f"open_worktree is not a directory: {open_worktree}")
    if not a_patterns:
        raise ValueError("at least one a_pattern is required")
    if not b_patterns:
        raise ValueError("at least one b_pattern is required")

    await git.create_branch(
        open_worktree, open_project, tracking=open_main_branch, skip_exist=True
    )
    await git.switch_branch(open_worktree, open_project)

    with open(os.path.join(open_worktree, "metadata.json"), "r") as f:
        metadata = json.load(f)

    for patch_name in metadata:
        patch_path = os.path.join(open_worktree, f"{patch_name}.patch")
        if not os.path.isfile(patch_path):
            continue

        with open(patch_path, "r", encoding="utf-8") as f:
            original_text = f.read()

        p = patch.Patch(os.path.basename(patch_path), original_text)
        p.pick_diff(a_patterns, b_patterns)
        result = p.render()

        if result != original_text:
            with open(patch_path, "w", encoding="utf-8") as f:
                f.write(result)

    del kwargs


async def pick_rename(
    a_patterns: list[re.Pattern[str]],
    b_patterns: list[re.Pattern[str]],
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    await _pick_rename(a_patterns, b_patterns, open_worktree=open_worktree, **kwargs)
    await git.commit(
        open_worktree,
        message="pick rename\n\n"
        + "\n".join(p.pattern for p in a_patterns)
        + "\n  ->\n"
        + "\n".join(p.pattern for p in b_patterns),
    )


async def pick_files(
    patterns: list[re.Pattern[str]],
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    await _pick_rename(patterns, patterns, open_worktree=open_worktree, **kwargs)
    await git.commit(
        open_worktree,
        message="pick files\n\n" + "\n".join(p.pattern for p in patterns),
    )


async def _remove_rename(
    a_patterns: list[re.Pattern[str]],
    b_patterns: list[re.Pattern[str]],
    *,
    open_worktree: str = "",
    open_project: str = "",
    open_main_branch: str = "",
    **kwargs,
) -> None:
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not os.path.isdir(open_worktree):
        raise ValueError(f"open_worktree is not a directory: {open_worktree}")
    if not a_patterns:
        raise ValueError("at least one a_pattern is required")
    if not b_patterns:
        raise ValueError("at least one b_pattern is required")

    await git.create_branch(
        open_worktree, open_project, tracking=open_main_branch, skip_exist=True
    )
    await git.switch_branch(open_worktree, open_project)

    with open(os.path.join(open_worktree, "metadata.json"), "r") as f:
        metadata = json.load(f)

    for patch_name in metadata:
        patch_path = os.path.join(open_worktree, f"{patch_name}.patch")
        if not os.path.isfile(patch_path):
            continue

        with open(patch_path, "r", encoding="utf-8") as f:
            original_text = f.read()

        p = patch.Patch(os.path.basename(patch_path), original_text)
        p.drop_diff(a_patterns, b_patterns)
        result = p.render()

        if result != original_text:
            with open(patch_path, "w", encoding="utf-8") as f:
                f.write(result)

    del kwargs


async def remove_rename(
    a_patterns: list[re.Pattern[str]],
    b_patterns: list[re.Pattern[str]],
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    await _remove_rename(a_patterns, b_patterns, open_worktree=open_worktree, **kwargs)
    await git.commit(
        open_worktree,
        message="remove rename\n\n"
        + "\n".join(p.pattern for p in a_patterns)
        + "\n  ->\n"
        + "\n".join(p.pattern for p in b_patterns),
    )


async def remove_files(
    patterns: list[re.Pattern[str]],
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    await _remove_rename(patterns, patterns, open_worktree=open_worktree, **kwargs)
    await git.commit(
        open_worktree,
        message="remove files\n\n" + "\n".join(p.pattern for p in patterns),
    )
