import asyncio
import logging
import os
import shlex
import tarfile
import tempfile

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init
import seed.devprod.starlark.inbazel.python.inbazel as inbazel

logger = logging.getLogger(__name__)

flag_out = seed_flag.define_string("out", "")
flag_subdir = seed_flag.define_string("subdir", "")
flag_strip_components = seed_flag.define_string("strip_components", "-1")
flag_local = seed_flag.define_bool("local", False)
arg_src_list = seed_flag.define_positional("src_list", "")


async def main():
    seed_init.initialize()
    srcs: list[str] = []
    with open(arg_src_list.get(), "r") as f:
        line = f.readline()
        while line:
            srcs.extend(shlex.split(line))
            line = f.readline()
    out = flag_out.get()
    strip_components = (
        int(flag_strip_components.get()) if flag_strip_components.get() else -1
    )
    unpack_folder = tempfile.mkdtemp()
    target_folder = os.path.join(unpack_folder, flag_subdir.get())
    inbazel.copy_srcs(
        srcs, target_folder, strip_components, sandbox=not flag_local.get()
    )
    # Sort directory and file entries at every level for deterministic builds.
    # os.walk and os.listdir return entries in OS-dependent order.
    with tarfile.open(out, "w:gz") as tar:
        for root, dirs, files in os.walk(unpack_folder):
            dirs.sort()  # in-place sort controls os.walk's traversal order
            for name in sorted(files):
                full_path = os.path.join(root, name)
                arcname = os.path.relpath(full_path, unpack_folder)
                tar.add(full_path, arcname=arcname, recursive=False)


if __name__ == "__main__":
    asyncio.run(main())
