"""Async wrappers over the git CLI used by the openfork tools."""

import asyncio

import seed.devprod.openfork.asyncx as asyncx


async def get_user_name(worktree: str) -> str:
    user_name = (
        await asyncx.output(["git", "config", "user.name"], cwd=worktree)
    ).strip()
    return user_name


async def get_user_email(worktree: str) -> str:
    user_email = (
        await asyncx.output(["git", "config", "user.email"], cwd=worktree)
    ).strip()
    return user_email


async def create_branch(
    worktree: str, branch: str, *, tracking: str = "main", skip_exist: bool = False
) -> None:
    if skip_exist:
        branch_list = await list_branches(worktree)
        if branch in branch_list:
            return
    await asyncx.run(
        ["git", "branch", "--track=direct", branch, tracking],
        cwd=worktree,
    )


async def delete_branch(worktree: str, branch: str) -> None:
    branch_list = await list_branches(worktree)
    if branch not in branch_list:
        return
    current_branch = (
        await asyncx.output(["git", "branch", "--show-current"], cwd=worktree)
    ).strip()
    if current_branch == branch:
        await asyncx.run(["git", "checkout", "--detach"], cwd=worktree)
    await asyncx.run(["git", "branch", "-D", branch], cwd=worktree)


async def list_branches(worktree: str) -> list[str]:
    return (
        await asyncx.output(
            ["git", "branch", "--list", "--format=%(refname:short)"],
            cwd=worktree,
        )
    ).splitlines()


async def switch_branch(worktree: str, branch: str) -> None:
    current_branch = (
        await asyncx.output(["git", "branch", "--show-current"], cwd=worktree)
    ).strip()
    if current_branch != branch:
        await asyncx.run(["git", "checkout", branch], cwd=worktree)
    worktree_status = (
        await asyncx.output(["git", "status", "--porcelain"], cwd=worktree)
    ).strip()
    if worktree_status:
        raise RuntimeError(f"worktree is not clean:\n{worktree_status}")


async def commit(worktree: str, *, message: str = "", skip_empty: bool = False) -> None:
    if not message:
        raise ValueError("commit message is required")
    await asyncx.run(["git", "add", "--all"], cwd=worktree)
    if skip_empty:
        worktree_status = (
            await asyncx.output(["git", "status", "--porcelain"], cwd=worktree)
        ).strip()
        if not worktree_status:
            return
    await asyncx.run(
        ["git", "commit", "--file", "-"],
        cwd=worktree,
        stdin=message,
    )


async def list_commits(worktree: str, branch: str) -> list[str]:
    # Uses raw asyncio because asyncx cannot suppress stderr or swallow the
    # non-zero returncode that git rev-list emits when `branch` does not
    # exist yet (e.g. on a fresh worktree).
    process = await asyncio.create_subprocess_exec(
        "git",
        "rev-list",
        "--reverse",
        branch,
        stdout=asyncio.subprocess.PIPE,
        stderr=asyncio.subprocess.DEVNULL,
        cwd=worktree,
    )
    stdout, _ = await process.communicate()
    if process.returncode:
        # Branch may not exist yet (e.g. fresh worktree); treat as empty.
        return []
    return stdout.decode().splitlines()
