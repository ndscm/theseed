# RBE Cache Proxy

`rbe-proxy-server` is a small HTTP server that speaks the Bazel HTTP remote
cache protocol (`GET`/`HEAD`/`PUT` on `/ac/<hash>` and `/cas/<hash>`) and
forwards it to a backend HTTP cache and/or a local disk cache. Its job is to let
Bazel reach the GCS cache bucket on a network where Bazel cannot reach it
directly.

## Why this exists

Bazel's remote cache client **cannot use an HTTP proxy**. It connects directly
to the `--remote_cache` host; the only "proxy" it supports is a Unix domain
socket (`--remote_proxy=unix:/path`). See
`CombinedCacheClientFactory.createHttp` in the Bazel source: a non-`unix:` proxy
is rejected outright, and there is no `http_proxy`/`https_proxy` handling
anywhere in the remote HTTP cache path.

On a proxy-only build network (all external egress via `HTTPS_PROXY`), that
means `--remote_cache=https://<bucket>.storage.googleapis.com/` times out —
Bazel dials GCS directly and never uses the proxy.

`rbe-proxy-server` bridges the gap. Bazel talks to it over a Unix socket
(unauthenticated, local), and the proxy reaches GCS over its HTTPS endpoint
using Go's default transport, which **does** honor `HTTPS_PROXY`. It keeps the
bucket in the raw `/cas`, `/ac` object layout, so the same bucket also works
with a direct `--remote_cache` configuration from a network that can reach GCS
(e.g. CI on GCP).

## Usage

Run `start.sh`. It builds the proxy, starts it on a Unix socket backed by
`~/.cache/bazel/remote`, and writes `rbe.local.bazelrc` (git-ignored, loaded via
`try-import`) pointing Bazel at the socket. The file and socket are removed on
exit.

```
seed/devprod/rbe/proxy/server/start.sh
```

Flags (env prefix `RBE_PROXY_`, falling back to `SEED_`):

- `--remote_cache` — backend HTTP cache base URL (e.g.
  `https://<bucket>.storage.googleapis.com/`). Optional.
- `--local_cache` — local disk cache directory, checked before the backend and
  populated from it. Optional.
- `--listen` — `[host]:port` or `unix:/path/to/socket` (matches Bazel's
  `--remote_proxy=unix:` form).
- `--google_default_credentials` — attach an ADC bearer token to backend
  requests. A client-supplied `Authorization` header is never overridden.

At least one of `--remote_cache` / `--local_cache` is required, giving three
modes: pure forwarder, read-through/write-through, or self-contained local
cache.

## Caveats

- **Bazel's `--google_default_credentials` must be disabled on the Bazel side.**
  The proxy socket is local and unauthenticated, and the proxy does the GCS
  auth. If Bazel is left with `--google_default_credentials` (the workspace
  default), it tries to fetch ADC and hangs on the GCE metadata server off-GCP
  ("Error initializing RemoteModule"). `rbe.local.bazelrc` sets
  `--google_default_credentials=false`.

- **`.bazelrc` override ordering.** `rbe.local.bazelrc` must be `try-import`ed
  _after_ the workspace's own `--remote_cache` default, or that later default
  wins and the proxy is bypassed. Bazel applies rc options in file order, and
  the last value of a single-valued flag takes effect.

- **GCS auth needs a service account, not user credentials.** With
  `--google_default_credentials`, the proxy uses ADC. User credentials
  (`gcloud auth application-default login`) are subject to org reauth policies
  and fail non-interactively with `invalid_rapt`. Point
  `GOOGLE_APPLICATION_CREDENTIALS` at a service-account key with
  `roles/storage.objectAdmin` on the bucket.

- **The local cache is not revalidated.** A local hit for `/ac/<hash>` is served
  without checking the backend. This is always safe for content-addressed
  `/cas/<hash>`, but the action cache is mutable, so a local `ac` entry wins
  until evicted. Same tradeoff as bazel-remote's local cache.

- **The Unix socket is owner-only.** `listen` binds it `0700` (by tightening the
  umask around the bind, not a racy post-bind `chmod`). The socket is
  unauthenticated but the proxy behind it holds ADC read/write to the shared
  bucket, so a world-connectable socket would let any local user read and poison
  the cache through the proxy's credentials. This is a single-user-workstation
  tool, so the perm is cheap defense-in-depth rather than a real multi-tenant
  boundary.

- **CAS contents are not verified.** On `PUT /cas/<digest>` the proxy tees the
  body straight to the backend without checking that it hashes to `<digest>`, so
  a buggy or malicious client can write mismatched content to the shared bucket.
  This is largely self-limiting — Bazel re-hashes CAS blobs on download and
  treats a mismatch as a cache miss, not a corrupt build — and matches the
  bazel-direct-to-GCS model, so it is low severity. A cheap sha256 check on the
  CAS write path would make the proxy poisoning-resistant on its own; it is left
  out for now to keep the write path a pure stream.

- **Bootstrapping.** `start.sh` writes `rbe.local.bazelrc` before building the
  proxy, so the build that produces the proxy is run with `--remote_cache=` to
  disable the (not-yet-running) cache for that one command.

## Alternatives considered

- **Bazel direct to GCS**
  (`--remote_cache=https://storage.googleapis.com/<bucket>`
  `--google_default_credentials`). Simplest — no server at all, and GCS's XML
  API _is_ the Bazel HTTP cache protocol. Rejected for the build network because
  Bazel can't traverse the HTTP proxy (see _Why this exists_). This is the right
  choice where GCS is directly reachable, and it shares the bucket with this
  proxy because both use the raw `/cas`, `/ac` layout.

- **bazel-remote** (buchgr/bazel-remote). Mature, with a local disk cache, a GCS
  proxy backend, and Unix-socket listening. Rejected as the primary because it
  stores objects under `cas.v2/`, `ac.v2/` in its own `casblob` format (not
  configurable), which is **not** interchangeable with Bazel-direct's raw layout
  — so the bucket could not be shared between the proxy path and direct GCS
  access. Still a reasonable option if that sharing is not needed.

- **A deployed cache server** (`../server` with `--storage=gcloud`, on Cloud Run
  or on-prem). Uses the GCS client library and preserves the raw layout. Cloud
  Run is a poor fit as a cache: its 32 MiB request limit rejects large CAS
  blobs, and clients on the build network reach it only through the same HTTP
  proxy Bazel can't use. An on-prem instance works, but this proxy is the
  lighter local-dev front: no service to operate, just a socket and a disk
  cache.
