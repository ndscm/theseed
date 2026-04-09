"""Replace hunk lines in every patch's matching file diffs by regex."""

import json
import logging
import os
import re

import seed.devprod.openfork.git as git
import seed.devprod.openfork.patch as patch

logger = logging.getLogger(__name__)


async def _replace_rename_hunk_lines(
    a_patterns: list[re.Pattern[str]],
    b_patterns: list[re.Pattern[str]],
    search_patterns: list[re.Pattern[str]],
    replace: str,
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
        p.replace_diff_hunk_lines(a_patterns, b_patterns, search_patterns, replace)
        result = p.render()

        if result != original_text:
            with open(patch_path, "w", encoding="utf-8") as f:
                f.write(result)

    del kwargs


async def replace_rename_hunk_lines(
    a_patterns: list[re.Pattern[str]],
    b_patterns: list[re.Pattern[str]],
    search_patterns: list[re.Pattern[str]],
    replace: str,
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    await _replace_rename_hunk_lines(
        a_patterns,
        b_patterns,
        search_patterns,
        replace,
        open_worktree=open_worktree,
        **kwargs,
    )
    await git.commit(
        open_worktree,
        message="replace rename hunk lines\n\n"
        + "\n".join(p.pattern for p in a_patterns)
        + "\n  ->\n"
        + "\n".join(p.pattern for p in b_patterns)
        + "\nsearch:\n  "
        + "\n  ".join(p.pattern for p in search_patterns)
        + "\nreplace:\n  "
        + replace,
    )


async def replace_hunk_lines(
    path_patterns: list[re.Pattern[str]],
    search_patterns: list[re.Pattern[str]],
    replace: str,
    *,
    open_worktree: str = "",
    **kwargs,
) -> None:
    await _replace_rename_hunk_lines(
        path_patterns,
        path_patterns,
        search_patterns,
        replace,
        open_worktree=open_worktree,
        **kwargs,
    )
    await git.commit(
        open_worktree,
        message="replace hunk lines\n\n"
        + "\n".join(p.pattern for p in path_patterns)
        + "\nsearch:\n  "
        + "\n  ".join(p.pattern for p in search_patterns)
        + "\nreplace:\n  "
        + replace,
    )
