"""Flags and accessors for the main, open, and rebuild worktree paths."""

import logging
import os

import seed.devprod.openfork.asyncx as asyncx
import seed.infra.python.seed_flag as seed_flag

logger = logging.getLogger(__name__)

flag_main_worktree = seed_flag.define_string("main_worktree", "")
flag_main_branch = seed_flag.define_string("main_branch", "origin/main")
flag_open_worktree = seed_flag.define_string("open_worktree", "/tmp/open")
flag_open_main_branch = seed_flag.define_string("open_main_branch", "main")
flag_open_project = seed_flag.define_string("open_project", "")
flag_rebuild_worktree = seed_flag.define_string("rebuild_worktree", "/tmp/rebuild")


async def get_main_worktree(arg_main_worktree: str = "") -> str:
    main_worktree = (arg_main_worktree or flag_main_worktree.get()).strip()
    if not main_worktree:
        raise ValueError("--main_worktree <main_worktree> is required")
    main_worktree = os.path.expanduser(main_worktree)
    if not os.path.isdir(main_worktree):
        raise ValueError(f"main_worktree is not a directory: {main_worktree}")
    return main_worktree


async def get_main_branch(arg_main_branch: str = "origin/main") -> str:
    main_branch = (arg_main_branch or flag_main_branch.get()).strip()
    if not main_branch:
        raise ValueError("--main_branch <main_branch> is required")
    return main_branch


async def get_open_worktree(
    arg_open_worktree: str = "",
    *,
    create: bool = False,
    open_main_branch: str = "",
    user_name: str = "",
    user_email: str = "",
) -> str:
    open_worktree = (arg_open_worktree or flag_open_worktree.get()).strip()
    if not open_worktree:
        raise ValueError("--open_worktree <open_worktree> is required")
    open_worktree = os.path.expanduser(open_worktree)
    if not os.path.isdir(open_worktree):
        if create:
            if not open_main_branch:
                open_main_branch = await get_open_main_branch()
            os.makedirs(open_worktree)
            await asyncx.run(
                ["git", "init", "--initial-branch", open_main_branch],
                cwd=open_worktree,
            )
            if user_name:
                await asyncx.run(
                    ["git", "config", "user.name", user_name],
                    cwd=open_worktree,
                )
            if user_email:
                await asyncx.run(
                    ["git", "config", "user.email", user_email],
                    cwd=open_worktree,
                )
        else:
            raise ValueError(f"open_worktree is not a directory: {open_worktree}")
    return open_worktree


async def get_open_main_branch(arg_open_main_branch: str = "main") -> str:
    open_main_branch = (arg_open_main_branch or flag_open_main_branch.get()).strip()
    if not open_main_branch:
        raise ValueError("--open_main_branch <open_main_branch> is required")
    return open_main_branch


async def get_open_project(arg_open_project: str = "") -> str:
    open_project = (arg_open_project or flag_open_project.get()).strip()
    if not open_project:
        raise ValueError("--open_project <open_project> is required")
    return open_project


async def get_rebuild_worktree(
    arg_rebuild_worktree: str = "/tmp/rebuild",
    *,
    create: bool = False,
    open_project: str = "",
    user_name: str = "",
    user_email: str = "",
) -> str:
    rebuild_worktree = (arg_rebuild_worktree or flag_rebuild_worktree.get()).strip()
    if not rebuild_worktree:
        raise ValueError("--rebuild_worktree <rebuild_worktree> is required")
    rebuild_worktree = os.path.expanduser(rebuild_worktree)
    if not os.path.isdir(rebuild_worktree):
        if create:
            if not open_project:
                open_project = await get_open_project()
            os.makedirs(rebuild_worktree)
            await asyncx.run(
                ["git", "init", "--initial-branch", f"rebuild/{open_project}"],
                cwd=rebuild_worktree,
            )
            if user_name:
                await asyncx.run(
                    ["git", "config", "user.name", user_name],
                    cwd=rebuild_worktree,
                )
            if user_email:
                await asyncx.run(
                    ["git", "config", "user.email", user_email],
                    cwd=rebuild_worktree,
                )
        else:
            raise ValueError(f"rebuild_worktree is not a directory: {rebuild_worktree}")
    return rebuild_worktree
