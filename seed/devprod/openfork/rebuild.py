"""Rebuild a git history from format-patch files."""

import json
import logging
import os

import seed.devprod.openfork.asyncx as asyncx
import seed.devprod.openfork.git as git

logger = logging.getLogger(__name__)


async def rebuild(
    termination: str = "",
    *,
    open_worktree: str = "",
    open_project: str = "",
    rebuild_worktree: str = "",
    fresh: bool = False,
    **kwargs,
) -> None:
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not os.path.isdir(open_worktree):
        raise ValueError(f"open_worktree is not a directory: {open_worktree}")
    if not open_project:
        raise ValueError("open_project is required")
    if not termination:
        raise ValueError("termination is required")
    if not rebuild_worktree:
        raise ValueError("rebuild_worktree is required")
    if not os.path.isdir(rebuild_worktree):
        raise ValueError(f"rebuild_worktree is not a directory: {rebuild_worktree}")

    with open(os.path.join(open_worktree, "metadata.json"), "r") as f:
        metadata: dict[str, dict[str, str]] = json.load(f)

    rebuild_metadata: dict[str, dict[str, str]] = {}
    if os.path.isfile(os.path.join(open_worktree, "rebuild.json")):
        with open(os.path.join(open_worktree, "rebuild.json"), "r") as f:
            rebuild_metadata = json.load(f)

    commit_long_indices: list[str] = [key for key in metadata if key <= (termination)]

    branch_list = await git.list_branches(rebuild_worktree)
    if f"rebuild/{open_project}" in branch_list:
        if fresh:
            await asyncx.run(
                ["git", "checkout", "--detach"],
                cwd=rebuild_worktree,
            )
            await asyncx.run(
                ["git", "branch", "-D", f"rebuild/{open_project}"],
                cwd=rebuild_worktree,
            )
    else:
        fresh = True

    if fresh:
        await asyncx.run(
            ["git", "checkout", "--orphan", f"rebuild/{open_project}"],
            cwd=rebuild_worktree,
        )
    else:
        await asyncx.run(
            ["git", "checkout", f"rebuild/{open_project}"],
            cwd=rebuild_worktree,
        )

    all_rebuilt_commits = await git.list_commits(
        rebuild_worktree, f"rebuild/{open_project}"
    )

    rebuilding = 0
    for commit_index in range(len(commit_long_indices)):
        patch_name = commit_long_indices[commit_index]
        commit_patch = os.path.abspath(
            os.path.join(open_worktree, f"{patch_name}.patch")
        )
        if not os.path.isfile(commit_patch):
            continue
        if rebuilding < len(all_rebuilt_commits):
            rebuilding += 1
            continue
        await asyncx.run(
            ["git", "am", "--empty=keep", commit_patch],
            cwd=rebuild_worktree,
        )
        commit_metadata = {
            **metadata[patch_name],
            **rebuild_metadata.get(patch_name, {}),
        }
        commit_message = commit_metadata.get("message")
        if commit_message:
            await asyncx.run(
                ["git", "commit", "--amend", "--no-edit", "--allow-empty", "--file=-"],
                cwd=rebuild_worktree,
                stdin=commit_message,
            )
        amend_env = os.environ.copy()
        amend_env["GIT_AUTHOR_NAME"] = commit_metadata["authorName"]
        amend_env["GIT_AUTHOR_EMAIL"] = commit_metadata["authorEmail"]
        amend_env["GIT_AUTHOR_DATE"] = commit_metadata["authorTime"]
        amend_env["GIT_COMMITTER_NAME"] = commit_metadata["committerName"]
        amend_env["GIT_COMMITTER_EMAIL"] = commit_metadata["committerEmail"]
        amend_env["GIT_COMMITTER_DATE"] = commit_metadata["committerTime"]
        await asyncx.run(
            [
                "git",
                "commit",
                "--amend",
                "--no-edit",
                "--allow-empty",
            ],
            cwd=rebuild_worktree,
            env=amend_env,
        )
        rebuilding += 1

    del kwargs
