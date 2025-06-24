import os
import typing

import fastmcp
import fastmcp.client


class McpClient:
    server: str
    service_name: str
    _client: fastmcp.Client[fastmcp.client.StreamableHttpTransport]

    def __init__(self, server: str = "", service_name: str = ""):
        self.service_name = service_name
        self.server = server
        if not self.server:
            self.server = os.environ.get("MCP_SERVER", "http://127.0.0.1:6277")  # MCPS
        self._client = typing.cast(
            fastmcp.Client[fastmcp.client.StreamableHttpTransport],
            fastmcp.Client(f"{server}/{service_name}/mcp/"),
        )

    def client(self) -> fastmcp.Client[fastmcp.client.StreamableHttpTransport]:
        return self._client
