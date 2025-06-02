import argparse
import asyncio
import os

import fastmcp


async def main():
    parser = argparse.ArgumentParser()
    parser.add_argument("service", type=str, nargs="?", default="")
    args = parser.parse_args()
    client = fastmcp.Client(
        os.environ.get("NEWTYPE_MCP_SERVER", "http://127.0.0.1:4364")
        + (f"/{args.service}" if args.service else "")
        + "/mcp/"
    )
    async with client:
        tools = await client.list_tools()
        for tool in tools:
            print(f"\x1b[1;36m{tool.name}:\x1b[0m\n{tool}\n")


if __name__ == "__main__":
    asyncio.run(main())
