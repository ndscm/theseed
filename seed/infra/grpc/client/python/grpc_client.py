import os

import grpc


class GrpcClient:
    sever: str

    def __init__(self, server):
        self.server = server
        if not self.server:
            self.server = os.environ.get("GRPC_SERVER", "127.0.0.1:4772")  # GRPC

    def channel(self):
        return grpc.aio.insecure_channel(self.server)
