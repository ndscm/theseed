import asyncio
import os

import fastmcp

import seed.infra.python.seed_flag as seed_flag
import seed.infra.python.seed_init as seed_init

arg_service = seed_flag.define_positional("service", "")


async def main():
    seed_init.initialize()
    service = arg_service.get()
    client = fastmcp.Client(
        os.environ.get("NEWTYPE_MCP_SERVER", "http://127.0.0.1:4364")
        + (f"/{service}" if service else "")
        + "/mcp/"
    )
    async with client:
        tools = await client.list_tools()
        for tool in tools:
            print(f"\x1b[1;36m{tool.name}:\x1b[0m\n{tool}\n")


if __name__ == "__main__":
    asyncio.run(main())
