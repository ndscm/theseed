import asyncio
import logging
import os
import shutil
import tempfile

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

logger = logging.getLogger(__name__)

flag_out = seed_flag.define_string("out", "")
flag_subdir = seed_flag.define_string("subdir", "")
arg_srcs = seed_flag.define_positional_list("srcs", [])


async def main():
    seed_init.initialize()
    srcs: list[str] = arg_srcs.get()
    out = flag_out.get()
    unpack_folder = tempfile.mkdtemp()
    target_folder = os.path.join(unpack_folder, flag_subdir.get())
    os.makedirs(target_folder, exist_ok=True)
    for src in srcs:
        if os.path.isdir(src):
            shutil.copytree(src, target_folder, dirs_exist_ok=True)
            # Revoke read-only permission from bazel
            os.chmod(target_folder, 0o755)
            for target_root, target_dirs, _ in os.walk(target_folder):
                for target_dir in target_dirs:
                    os.chmod(os.path.join(target_root, target_dir), 0o755)
            continue
        shutil.copy(src, target_folder)
    output_folder = tempfile.mkdtemp()
    output_archive_path = shutil.make_archive(
        os.path.join(output_folder, "out"), "gztar", unpack_folder
    )
    shutil.move(output_archive_path, out)


if __name__ == "__main__":
    asyncio.run(main())
