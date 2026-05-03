# Seed Front End (SFE)

SFE is a high-performance reverse proxy built around the concept of code as
configuration. You only pull the components you need into your binary. With this
approach, SFE can achieve 10x the throughput of nginx for a simple TLS
termination reverse proxy use case.

The intended way to use SFE is to reference the SFE codebase as a skeleton and
write your own version of xxxFE.

Like any reverse proxy, SFE handles these common responsibilities:

- Authenticate users and manage sessions
- Unwrap long-lived sessions into short-lived JWTs, so backing services can
  easily identify users and enforce permissions
- Manage external-facing certificates and terminate the external TLS layer
- Dynamically rotate HTTPS certificates
- Dynamically route by domain for multi-tenancy platforms

SFE has three common deployment patterns:

- Centralized reverse proxy for your entire cluster, handling domain-based
  routing.
- Sidecar for an individual service, adding features like auth guards.
- Melt SFE into your server for the best possible performance.
