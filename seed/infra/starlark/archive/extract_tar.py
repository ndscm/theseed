import asyncio
import logging
import os
import shutil

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

logger = logging.getLogger(__name__)

flag_out = seed_flag.define_string("out", "")
arg_srcs = seed_flag.define_positional_list("srcs", [])


async def main():
    seed_init.initialize()
    srcs: list[str] = arg_srcs.get()
    out = flag_out.get()
    os.makedirs(out, exist_ok=True)
    for src in srcs:
        if src.endswith(".tar.gz"):
            shutil.unpack_archive(src, out)
            continue


if __name__ == "__main__":
    asyncio.run(main())
