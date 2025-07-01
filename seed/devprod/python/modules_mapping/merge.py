import asyncio
import json

import pydantic_settings

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

arg_srcs = seed_flag.define("srcs", pydantic_settings.CliPositionalArg[list[str]], [])


async def main():
    seed_init.initialize()
    srcs: list[str] = arg_srcs.get()
    merged: dict = {}
    for src in srcs:
        with open(src, "r") as f:
            mapping = json.load(f)
            merged.update(mapping)
    print(json.dumps(merged, indent=2))


if __name__ == "__main__":
    asyncio.run(main())
