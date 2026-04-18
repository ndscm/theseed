import asyncio
import logging
import re
import subprocess

import seed.devprod.openfork.clean as clean
import seed.devprod.openfork.diffedit as diffedit
import seed.devprod.openfork.diffpath as diffpath
import seed.devprod.openfork.export as export
import seed.devprod.openfork.git as git
import seed.devprod.openfork.patch as patch
import seed.devprod.openfork.prepare as prepare
import seed.devprod.openfork.rebuild as rebuild
import seed.devprod.openfork.walk as walk
import seed.devprod.openfork.worktree as worktree
import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

logger = logging.getLogger(__name__)

flag_export = seed_flag.define_bool("export", True)


async def build(**kwargs) -> None:

    await diffpath.pick_files(
        [
            re.compile(r"^.bazeliskrc$"),
            re.compile(r"^.bazelrc$"),
            re.compile(r"^.gitignore$"),
            re.compile(r"^MODULE.bazel$"),
            re.compile(r"^WORKSPACE.bazel$"),
            re.compile(r"^go.mod$"),
            re.compile(r"^go.sum$"),
        ]
        + [
            re.compile(r"^seed/devprod/buildinfo/generate-workspace-status.sh$"),
            re.compile(r"^seed/devprod/ndscm/.*$"),
            re.compile(r"^seed/infra/init/go/.*$"),
            re.compile(r"^seed/infra/log/go/.*$"),
            re.compile(r"^seed/infra/error/go/.*$"),
            re.compile(r"^seed/infra/dotenv/go/.*$"),
            re.compile(r"^seed/infra/shell/go/.*$"),
            re.compile(r"^seed/infra/flag/go/.*$"),
        ],
        **kwargs,
    )
    await diffedit.replace_hunk_lines(
        [re.compile(r"^MODULE.bazel$")],
        [
            re.compile(r"^.*\]\,.*$"),
            re.compile(r"^.*Android.*$"),
            re.compile(r"^.*android.*$"),
            re.compile(r"^.*artifact.*$"),
            re.compile(r"^.*aspect.*$"),
            re.compile(r"^.*boost.*$"),
            re.compile(r"^.*CC.*$"),
            re.compile(r"^.*commit.*$"),
            re.compile(r"^.*Container.*$"),
            re.compile(r"^.*cpp.*$"),
            re.compile(r"^.*debian.*$"),
            re.compile(r"^.*docker.*$"),
            re.compile(r"^.*git_override.*$"),
            re.compile(r"^.*git_repository.*$"),
            re.compile(r"^.*hedron.*$"),
            re.compile(r"^.*http_archive.*$"),
            re.compile(r"^.*java.*$"),
            re.compile(r"^.*JavaScript.*$"),
            re.compile(r"^.*linux/amd64.*$"),
            re.compile(r"^.*maven.*$"),
            re.compile(r"^.*node.*$"),
            re.compile(r"^.*npm.*$"),
            re.compile(r"^.*oci.*$"),
            re.compile(r"^.*patches.*$"),
            re.compile(r"^.*pip.*$"),
            re.compile(r"^.*Pkg.*$"),
            re.compile(r"^.*pnpm.*$"),
            re.compile(r"^.*pypi.*$"),
            re.compile(r"^.*python.*$"),
            re.compile(r"^.*Python.*$"),
            re.compile(r"^.*requirements.*$"),
            re.compile(r"^.*rules_cc.*$"),
            re.compile(r"^.*rules_foreign_cc.*$"),
            re.compile(r"^.*rules_js.*$"),
            re.compile(r"^.*rules_jvm.*$"),
            re.compile(r"^.*rules_nodejs.*$"),
            re.compile(r"^.*rules_pkg.*$"),
            re.compile(r"^.*rules_ts.*$"),
            re.compile(r"^.*single_version_override.*$"),
            re.compile(r"^.*slim.*$"),
        ],
        "",
        **kwargs,
    )

    async def walk_module_bazel(n, p: patch.Patch, m) -> tuple[patch.Patch, None]:
        del n, m
        for diff in p.diffs:
            if diff.a == "MODULE.bazel" and diff.b == "MODULE.bazel":
                for i, line in enumerate(diff.hunk_lines):
                    if line == " )" and diff.hunk_lines[i - 1] in ["-", "+", " "]:
                        diff.hunk_lines[i] = " "
            if diff.a == "MODULE.bazel":
                for i, line in enumerate(diff.hunk_lines):
                    if line == "-)" and diff.hunk_lines[i - 1] in ["-", "+", " "]:
                        diff.hunk_lines[i] = "-"
            if diff.b == "MODULE.bazel":
                for i, line in enumerate(diff.hunk_lines):
                    if line == "+)" and diff.hunk_lines[i - 1] in ["-", "+", " "]:
                        diff.hunk_lines[i] = "+"
        return p, None

    await walk.walk(walk_module_bazel, **kwargs)

    await diffedit.replace_hunk_lines(
        [re.compile(r"^.bazelrc$")],
        [
            re.compile(r"^.*aspect.*$"),
            re.compile(r"^.*boost.*$"),
            re.compile(r"^.*cxx.*$"),
            re.compile(r"^.*parse_headers.*$"),
            re.compile(r"^.*rules_js.*$"),
        ],
        "",
        **kwargs,
    )


async def run(**kwargs) -> None:
    if flag_export.get():
        await export.export(**kwargs)

    await prepare.prepare(**kwargs)

    await build(**kwargs)

    await clean.clean_empty(**kwargs)

    # Stops the rebuild at commit index 580 (2026-04-18), update the script
    # and check results before rebuilding more commits.
    await rebuild.rebuild("10000580", fresh=True, **kwargs)
    subprocess.run(
        ["git", "tag", "initial"],
        check=True,
        cwd=kwargs["rebuild_worktree"],
    )


async def main() -> None:
    seed_init.initialize()

    main_worktree = await worktree.get_main_worktree("~/ndscm/theseed/main")
    main_branch = await worktree.get_main_branch()
    user_name = await git.get_user_name(main_worktree)
    user_email = await git.get_user_email(main_worktree)
    open_worktree = await worktree.get_open_worktree(
        "/tmp/open",
        create=True,
        user_name=user_name,
        user_email=user_email,
    )
    open_main_branch = await worktree.get_open_main_branch()
    open_project = await worktree.get_open_project("ndscm")
    rebuild_worktree = await worktree.get_rebuild_worktree(
        create=True,
        open_project=open_project,
        user_name=user_name,
        user_email=user_email,
    )

    await run(
        main_worktree=main_worktree,
        main_branch=main_branch,
        open_worktree=open_worktree,
        open_main_branch=open_main_branch,
        open_project=open_project,
        rebuild_worktree=rebuild_worktree,
    )


if __name__ == "__main__":
    asyncio.run(main())
