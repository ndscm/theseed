# vpcdirect

An `http.RoundTripper` that redirects outbound requests for a **public host** to
an **internal upstream**, while preserving the public host for the upstream's
own URL construction.

## Motivation

A service sometimes must call an endpoint by its public hostname — because that
host is baked into tokens, issuers, discovery documents, or config — at a time
when the public path is not usable. Two cases drive this:

- **Bootstrap.** The public endpoint depends on the very thing being set up.
  When a service terminates TLS for a public host and proxies it to an upstream,
  that same service calling the public host at startup loops back to itself
  before its certificate exists — a deadlock.

- **VPC-only reachability.** The public path exists but is not reachable from
  where the caller runs; only the internal (VPC) endpoint is.

In both cases the request must go to the internal upstream, yet still _present_
the public host so the upstream emits the right issuer/URLs. `vpcdirect` sends
the request to the upstream's scheme+host (so TCP, SNI, and TLS verification use
the upstream's own certificate) while moving the original host to
`X-Forwarded-Host`. This mirrors, on the client side, what a reverse proxy does
on the server side — and it relies on the upstream honoring `X-Forwarded-*`,
otherwise the upstream would build URLs for its own host rather than the public
one.
