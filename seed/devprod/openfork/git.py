"""Async wrappers over the git CLI used by the openfork tools."""

import seed.devprod.openfork.asyncx as asyncx


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
