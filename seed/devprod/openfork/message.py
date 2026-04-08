"""Rewrite per-patch commit messages via a callback, persisted to rebuild.json."""

import collections.abc
import email
import email.policy
import json
import os
import re

import seed.devprod.openfork.git as git

Callable = collections.abc.Callable
Awaitable = collections.abc.Awaitable


def read_commit_message(patch_file_path):
    with open(patch_file_path, "r", encoding="utf-8") as f:
        msg = email.message_from_file(f, policy=email.policy.default)
    subject = msg.get("Subject", "")
    commit_message = re.sub(r"^\[PATCH.*?\]\s*", "", subject)
    payload = str(msg.get_payload())
    commit_body_lines = []
    for line in payload.splitlines():
        if line == "---" or line.startswith("diff --git"):
            break
        commit_body_lines.append(line)
    commit_body = "\n".join(commit_body_lines).strip()
    if commit_body:
        commit_message += f"\n\n{commit_body}"
    return commit_message


async def update_messages(
    updater: Callable[[str, str], Awaitable[str]] | None = None,
    *,
    open_worktree: str,
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
        commit_message = rebuild_metadata.get(patch_name, {}).get("message")
        if not commit_message:
            patch_path = os.path.join(open_worktree, f"{patch_name}.patch")
            if not os.path.isfile(patch_path):
                continue
            commit_message = read_commit_message(patch_path)
        updated_message = await updater(patch_name, commit_message)
        if updated_message and updated_message != commit_message:
            rebuild_metadata[patch_name] = {
                **rebuild_metadata.get(patch_name, {}),
                "message": updated_message,
            }
    with open(os.path.join(open_worktree, f"rebuild.json"), "w") as f:
        json.dump(rebuild_metadata, f, indent=4)

    await git.commit(open_worktree, message="update messages\n\n" + updater.__name__)

    del kwargs
