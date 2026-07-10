# Bidirectional Request Service

gRPC unary requests work well for stateless microservices deployed in a cluster,
but there's no standard way for the server to send requests back to the client.
In theory, you can create a single TCP tunnel, attach a multiplexer to it, and
use the multiplexer to provide TCP streams for HTTP clients. Different
multiplexer libraries take different approaches to setting up the tunnel. What
we want is a stream-based reverse multiplexer that plays nicely with the HTTP
mux, so connections can traverse NATs, HTTP proxies, and reverse proxies.

## Design

Since we need to attach the multiplexer to the HTTP mux, we need an HTTP-based
bidirectional stream. WebSockets and gRPC streams are the two obvious choices.
We originally went with a gRPC bidirectional stream, but gRPC streams require
end-to-end HTTP/2, and they break behind proxies that only speak HTTP/1.1 (for
example, Cloud Run domain mapping cannot keep HTTP/2 end to end). WebSockets
start as a plain HTTP/1.1 upgrade request, so they traverse such proxies.

We switched the transport to WebSockets while keeping the rest of the design:

- The protobuf `Payload` message is still the wire format. Each payload is
  marshaled and sent as a single binary WebSocket frame.
- A `Payload` carries a _chunk_ of an HTTP/1.1 message rather than a whole one.
  A stream is a pair of independent byte streams, one per direction; each side
  ends its own direction with `end`, and either may abandon both with `reset`.
  Bodies are chunk-encoded and framed as they are produced, so a
  server-streaming response reaches the caller as the handler writes it, and a
  caller can keep sending request body while reading the response.
- The endpoint keeps the gRPC-style route path
  (`/seed.cloud.bidirequest.proto.BidirequestService/Connect`), so the proto
  service definition still names the route.
- The existing authorization infrastructure is still reused: the WebSocket
  handshake is a regular HTTP request, so the handler can be wrapped with the
  same HTTP middleware (OpenID JWT interceptor, bearer token middleware) as the
  gRPC handlers on the same mux. Note that the middleware runs once at handshake
  time, not per message: every `Payload` on the open connection inherits the
  auth context established at the handshake, unlike gRPC unary calls where each
  request is authorized independently.

## HTTP/2 and bidi streams

Nothing here puts HTTP/2 on the wire — the WebSocket is an HTTP/1.1 upgrade, and
the frames carry HTTP/1.1 messages.

connect refuses a bidirectional stream unless `ProtoMajor >= 2`, at both the
handler (`request.ProtoMajor`) and the caller (`response.ProtoMajor`). What that
check is really asking is whether the connection is full duplex. A stream here
is: the peer can send body frames while the handler is already sending response
frames. So after parsing a tunneled message, `MuxTransport` sets
`ProtoMajor = 2` on the in-memory `http.Request` and `http.Response` before
handing it on. Those fields are never serialized; no proxy between the peers
ever sees them.

This is what lets a `rpc Foo(stream A) returns (stream B)` run over the tunnel.
