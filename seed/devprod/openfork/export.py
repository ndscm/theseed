"""Export the main worktree's history as format-patch files in the open worktree."""

import json
import logging
import os
import re

import seed.devprod.openfork.asyncx as asyncx
import seed.devprod.openfork.git as git

logger = logging.getLogger(__name__)


async def export(
    *,
    main_worktree: str = "",
    main_branch: str = "origin/main",
    open_worktree: str = "",
    open_main_branch: str = "main",
    **kwargs,
) -> None:
    if not main_worktree:
        raise ValueError("main_worktree is required")
    if not os.path.isdir(main_worktree):
        raise ValueError(f"main_worktree is not a directory: {main_worktree}")
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not os.path.isdir(open_worktree):
        raise ValueError(f"open_worktree is not a directory: {open_worktree}")

    all_main_commits = await git.list_commits(main_worktree, main_branch)
    logger.info("All main commits: (%s)", len(all_main_commits))

    await git.switch_branch(open_worktree, open_main_branch)

    # We want to keep the patch file names stable, so we use a fixed start
    # number and filename max length to cut at the patch index. The result
    # patch file name will be like 10000001.patch, 10000002.patch, ...,
    # 10000010.patch, etc. This way we can avoid renaming patch files when the
    # commit history changes.
    all_main_patches = (
        await asyncx.output(
            [
                "git",
                "format-patch",
                "--no-stat",
                "--no-signature",
                "--root",
                main_branch,
                "--output-directory",
                open_worktree,
                "--start-number",
                "10000001",
                "--filename-max-length",
                "15",
            ],
            cwd=main_worktree,
        )
    ).splitlines()
    logger.info("All main patches: (%s)", len(all_main_patches))

    main_metadata = {}
    for commit_index in range(len(all_main_commits)):
        patch_name = str(10000001 + commit_index)
        target_commit = all_main_commits[commit_index]
        target_commit_file = await asyncx.output(
            ["git", "cat-file", "commit", target_commit],
            cwd=main_worktree,
        )
        author_match = re.search(
            r"^author (.*) [<](.*)[>] ([0-9]* [+][0-9]{4})$",
            target_commit_file,
            re.MULTILINE,
        )
        if not author_match:
            raise RuntimeError("no author match")
        author_name, author_email, author_time = author_match.groups()
        committer_match = re.search(
            r"^committer (.*) [<](.*)[>] ([0-9]* [+][0-9]{4})$",
            target_commit_file,
            re.MULTILINE,
        )
        if not committer_match:
            raise RuntimeError("no committer match")
        committer_name, committer_email, committer_time = committer_match.groups()
        main_metadata[patch_name] = {
            "commit": target_commit,
            "authorName": author_name,
            "authorEmail": author_email,
            "authorTime": author_time,
            "committerName": committer_name,
            "committerEmail": committer_email,
            "committerTime": committer_time,
        }
    with open(os.path.join(open_worktree, "metadata.json"), "w") as f:
        json.dump(main_metadata, f, indent=4)

    await git.commit(open_worktree, message="export", skip_empty=True)

    del kwargs
