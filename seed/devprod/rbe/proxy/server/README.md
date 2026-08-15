# RBE Cache Proxy

`rbe-proxy-server` is a small HTTP server that speaks the Bazel HTTP remote
cache protocol (`GET`/`HEAD`/`PUT` on `/ac/<hash>` and `/cas/<hash>`) and
forwards it to a backend HTTP cache and/or a local disk cache. Its job is to let
Bazel reach the GCS cache bucket on a network where Bazel cannot reach it
directly.

## Why this exists

Pointing every Bazel client at the cloud build cache (a GCS bucket) directly
means every action-cache lookup and CAS upload crosses the WAN to Google. On an
on-prem build network with many clients that is a lot of egress traffic, plus a
round-trip to the cloud on the critical path of every cache hit.

`rbe-proxy-server` is meant to run on-prem, close to the clients, as a
read-through/write-through front for the cloud cache. Bazel talks to a nearby
instance over its HTTP cache protocol; the proxy serves hits from its local disk
cache and only reaches GCS on a miss, so repeated fan-out from many clients
collapses into one shared local cache — less cloud egress and lower latency on
the hot path. The proxy reaches GCS over its HTTPS endpoint using Go's default
transport, which honors `HTTPS_PROXY` if the on-prem network needs one to leave.
It keeps the bucket in the raw `/cas`, `/ac` object layout, so the same bucket
also works with a direct `--remote_cache` configuration from a network that can
reach GCS directly (e.g. CI on GCP).

## Usage

Build and run the binary directly:

```
bazel run //seed/devprod/rbe/proxy/server -- \
  --remote_cache=https://<bucket>.storage.googleapis.com/ \
  --local_cache=/var/cache/rbe-proxy \
  --listen=:8080 \
  --credential_helper=/path/to/credential-helper \
  --credential_helper_format=oauth2
```

Then point Bazel at it with `--remote_cache=http://<host>:<port>/`, or with
`--remote_proxy=unix:/path/to/socket` when listening on a Unix socket.

Flags (env prefix `RBE_PROXY_`, falling back to `SEED_`):

- `--remote_cache` — backend HTTP cache base URL (e.g.
  `https://<bucket>.storage.googleapis.com/`). Optional.
- `--local_cache` — local disk cache directory, checked before the backend and
  populated from it. Optional.
- `--local_put` — accept uploads with no `Authorization` header by writing them
  only to the local cache, never forwarding them to the backend (requires
  `--local_cache`). When unset, such uploads are rejected.
- `--listen` — `[host]:port` or `unix:/path/to/socket` (matches Bazel's
  `--remote_proxy=unix:` form).
- `--credential_helper` — command run to mint a backend bearer token, attached
  to backend _reads_ (`GET`/`HEAD`) when the client sends no `Authorization`
  header. A client-supplied header is never overridden. Empty leaves the backend
  unauthenticated except for a client-supplied header.
- `--credential_helper_format` — how to read the helper's stdout: empty uses it
  as the bearer token verbatim (re-run every request, since a raw token carries
  no expiry); `oauth2` parses it as a JSON `oauth2.Token` and reuses the access
  token until it expires, then re-runs the helper to refresh.

At least one of `--remote_cache` / `--local_cache` is required, giving three
modes: pure forwarder, read-through/write-through, or self-contained local
cache.

## Authentication model

The proxy authenticates to the backend asymmetrically, so that an
unauthenticated client can read the shared cache but never write to it with the
proxy's credentials:

- **Reads (`GET`/`HEAD`).** A client `Authorization` header is forwarded
  unchanged. If the client sends none, the proxy attaches a backend bearer token
  from its `--credential_helper`, so unauthenticated clients still get
  read-through hits from the backend. This keeps the backend credential inside
  the proxy's trust domain — untrusted callers (CI jobs, dev machines) read the
  cache without ever holding the key.
- **Writes (`PUT`).** The proxy never lends its credential-helper token to a
  write. An upload must carry the client's own `Authorization`; without it the
  proxy forwards nothing to the backend, so an unauthenticated caller cannot
  poison the shared cache. With `--local_put` such an upload is instead confined
  to the local cache (never forwarded), so a local build can still populate the
  read-through cache; without it the upload is rejected with `403`. Only
  trusted, authenticated writers populate the backend. (Local-only mode has no
  backend credential to protect and always writes locally.)

Without `--local_put`, unauthenticated clients will see their cache uploads fail
with `403`; set `--remote_upload_local_results=false` on those clients to skip
the attempt and avoid the warning noise.

## Caveats

- **Bazel's `--google_default_credentials` must be disabled on the Bazel side.**
  The proxy endpoint is unauthenticated, and the proxy does the GCS auth. If
  Bazel is left with `--google_default_credentials` (the workspace default), it
  tries to fetch ADC and hangs on the GCE metadata server off-GCP ("Error
  initializing RemoteModule"). Set `--google_default_credentials=false` on the
  Bazel side.

- **`.bazelrc` override ordering.** The `--remote_cache` that points Bazel at
  the proxy must be applied _after_ the workspace's own `--remote_cache`
  default, or that later default wins and the proxy is bypassed. Bazel applies
  rc options in file order, and the last value of a single-valued flag takes
  effect.

- **The backend token comes from `--credential_helper`.** The proxy runs the
  configured helper to mint the backend bearer token (with
  `--credential_helper_format=oauth2` it reuses the token until it expires, then
  re-runs the helper to refresh). For a GCS backend the helper must yield a
  token for an identity with `roles/storage.objectAdmin` on the bucket. Avoid
  user credentials (`gcloud auth application-default login`): they are subject
  to org reauth policies and fail non-interactively with `invalid_rapt`. Use a
  service account or workload-identity federation instead.

- **The local cache is not revalidated.** A local hit for `/ac/<hash>` is served
  without checking the backend. This is always safe for content-addressed
  `/cas/<hash>`, but the action cache is mutable, so a local `ac` entry wins
  until evicted. Same tradeoff as bazel-remote's local cache.

- **The Unix socket is owner-only.** `listen` binds it `0700` (by tightening the
  umask around the bind, not a racy post-bind `chmod`). The socket is
  unauthenticated, and the proxy lends its ADC token to reads, so a
  world-connectable socket would let any local user read the shared bucket
  through the proxy. Writes still require the caller's own `Authorization`, so
  the proxy's credential cannot be used to poison the cache.

- **CAS contents are not verified.** On `PUT /cas/<digest>` the proxy tees the
  body straight to the backend without checking that it hashes to `<digest>`, so
  a buggy or malicious client can write mismatched content to the shared bucket.
  This is largely self-limiting — Bazel re-hashes CAS blobs on download and
  treats a mismatch as a cache miss, not a corrupt build — and matches the
  bazel-direct-to-GCS model, so it is low severity. A cheap sha256 check on the
  CAS write path would make the proxy poisoning-resistant on its own; it is left
  out for now to keep the write path a pure stream.

## Alternatives considered

- **Bazel direct to GCS**
  (`--remote_cache=https://storage.googleapis.com/<bucket>`
  `--google_default_credentials`). Simplest — no server at all, and GCS's XML
  API _is_ the Bazel HTTP cache protocol. This is the right choice where the WAN
  round-trip to GCS is acceptable, but every client then hits the cloud on every
  lookup with no shared local cache in front — which is what this proxy exists
  to avoid (see _Why this exists_). It shares the bucket with this proxy because
  both use the raw `/cas`, `/ac` layout.

- **bazel-remote** (buchgr/bazel-remote). Mature, with a local disk cache, a GCS
  proxy backend, and Unix-socket listening. Rejected as the primary because it
  stores objects under `cas.v2/`, `ac.v2/` in its own `casblob` format (not
  configurable), which is **not** interchangeable with Bazel-direct's raw layout
  — so the bucket could not be shared between the proxy path and direct GCS
  access. Still a reasonable option if that sharing is not needed.

- **A deployed cache server** (`../server` with `--storage=gcloud`, on Cloud Run
  or on-prem). Uses the GCS client library and preserves the raw layout. Cloud
  Run is a poor fit as a cache: its 32 MiB request limit rejects large CAS
  blobs, and being cloud-hosted it puts the WAN round-trip back on every lookup
  that this proxy is meant to remove. An on-prem instance works, but this proxy
  is the lighter front: no service to operate, just a socket and a disk cache.
