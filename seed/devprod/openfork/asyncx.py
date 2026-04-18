"""Async subprocess helpers used in place of blocking subprocess.run."""

import asyncio
import subprocess


async def run(
    args: list[str],
    *,
    cwd: str | None = None,
    stdin: str | None = None,
    env: dict[str, str] | None = None,
) -> None:
    # stderr is intentionally left inherited so the child's stderr streams
    # live to the parent process's stderr (e.g. the user's terminal).
    process = await asyncio.create_subprocess_exec(
        *args,
        stdin=asyncio.subprocess.PIPE if stdin is not None else None,
        cwd=cwd,
        env=env,
    )
    try:
        if stdin is not None:
            await process.communicate(input=stdin.encode())
        else:
            await process.wait()
    except BaseException:
        process.kill()
        await process.wait()
        raise
    if process.returncode:
        raise subprocess.CalledProcessError(process.returncode, list(args))


async def output(
    args: list[str],
    *,
    cwd: str | None = None,
    stdin: str | None = None,
    env: dict[str, str] | None = None,
) -> str:
    # stderr is intentionally left inherited so the child's stderr streams
    # live to the parent process's stderr (e.g. the user's terminal).
    process = await asyncio.create_subprocess_exec(
        *args,
        stdin=asyncio.subprocess.PIPE if stdin is not None else None,
        stdout=asyncio.subprocess.PIPE,
        cwd=cwd,
        env=env,
    )
    try:
        if stdin is not None:
            stdout, _ = await process.communicate(input=stdin.encode())
        else:
            stdout, _ = await process.communicate()
    except BaseException:
        process.kill()
        await process.wait()
        raise
    if process.returncode:
        raise subprocess.CalledProcessError(process.returncode, list(args))
    return stdout.decode()
