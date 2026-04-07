"""Reset the open project branch to a fresh copy of the open main branch."""

import logging

import seed.devprod.openfork.git as git

logger = logging.getLogger(__name__)


async def prepare(
    *,
    open_worktree: str = "",
    open_main_branch: str = "main",
    open_project: str = "",
    **kwargs,
) -> None:
    if not open_worktree:
        raise ValueError("open_worktree is required")
    if not open_project:
        raise ValueError("open_project is required")
    if not open_main_branch:
        raise ValueError("open_main_branch is required")

    await git.switch_branch(open_worktree, open_main_branch)
    await git.delete_branch(open_worktree, open_project)
    await git.create_branch(open_worktree, open_project, tracking=open_main_branch)
    await git.switch_branch(open_worktree, open_project)

    del kwargs
