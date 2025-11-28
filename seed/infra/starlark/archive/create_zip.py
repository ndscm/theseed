import asyncio
import logging
import os
import shutil
import tempfile

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init
import seed.infra.starlark.archive.inbazel as inbazel

logger = logging.getLogger(__name__)

flag_out = seed_flag.define_string("out", "")
flag_subdir = seed_flag.define_string("subdir", "")
flag_strip_components = seed_flag.define_string("strip_components", "-1")
arg_srcs = seed_flag.define_positional_list("srcs", [])


async def main():
    seed_init.initialize()
    srcs: list[str] = arg_srcs.get()
    out = flag_out.get()
    strip_components = (
        int(flag_strip_components.get()) if flag_strip_components.get() else -1
    )
    unpack_folder = tempfile.mkdtemp()
    target_folder = os.path.join(unpack_folder, flag_subdir.get())
    inbazel.copy_srcs(srcs, target_folder, strip_components)
    output_folder = tempfile.mkdtemp()
    output_archive_path = shutil.make_archive(
        os.path.join(output_folder, "out"), "zip", unpack_folder
    )
    shutil.move(output_archive_path, out)


if __name__ == "__main__":
    asyncio.run(main())
