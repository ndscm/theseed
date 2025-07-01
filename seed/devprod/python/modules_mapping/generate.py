import asyncio
import json
import platform

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

flag_build_workspace_directory = seed_flag.define(
    "build_workspace_directory", str, ""
)


async def main():
    seed_init.initialize()
    build_workspace_directory = flag_build_workspace_directory.get().strip()
    if not build_workspace_directory:
        raise ValueError("must run with bazel")
    with open("seed/devprod/python/modules_mapping/local_modules_mapping.json", "r") as f:
        mapping = json.load(f)
    if platform.system() == "Linux":
        output = f"{build_workspace_directory}/seed/devprod/python/modules_mapping/modules_mapping_linux.json"
    elif platform.system() == "Darwin":
        output = f"{build_workspace_directory}/seed/devprod/python/modules_mapping/modules_mapping_darwin.json"
    else:
        raise ValueError(f"not supported platform: {platform.system()}")
    with open(output, "w") as f:
        json.dump(mapping, f, indent=2)


if __name__ == "__main__":
    asyncio.run(main())
