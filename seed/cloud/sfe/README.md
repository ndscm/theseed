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

## Port 80 and 443

Rootless Podman can't bind to ports 80 or 443 by default. To work around this,
use a lightweight reverse proxy like HAProxy to forward port 443 to SFE's
default port 9443. Install HAProxy with `sudo dnf install haproxy`, then replace
`/etc/haproxy/haproxy.cfg` with the following:

```cfg
frontend http
    bind *:80
    mode tcp
    default_backend sfe_http

backend sfe_http
    mode tcp
    server sfe_http 127.0.0.1:9080

frontend https
    bind *:443
    mode tcp
    default_backend sfe_https

backend sfe_https
    mode tcp
    server sfe_https 127.0.0.1:9443
```

On SELinux-enabled systems, you'll also need to allow HAProxy to connect to
arbitrary ports:

```bash
sudo setsebool -P haproxy_connect_any 1
```
