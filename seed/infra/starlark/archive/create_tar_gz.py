import asyncio
import logging
import os
import shutil
import tempfile

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

logger = logging.getLogger(__name__)

flag_out = seed_flag.define_string("out", "")
arg_srcs = seed_flag.define_positional_list("srcs", [])


async def main():
    seed_init.initialize()
    srcs: list[str] = arg_srcs.get()
    out = flag_out.get()
    unpack_folder = tempfile.mkdtemp()
    for src in srcs:
        if os.path.isdir(src):
            shutil.copytree(src, unpack_folder, dirs_exist_ok=True)
            continue
        shutil.copy(src, unpack_folder)
    output_folder = tempfile.mkdtemp()
    output_archive_path = shutil.make_archive(
        os.path.join(output_folder, "out"), "gztar", unpack_folder
    )
    shutil.move(output_archive_path, out)


if __name__ == "__main__":
    asyncio.run(main())
