#!/usr/bin/env bash
# Start the local rbe-proxy-server that fronts the GCS cache bucket and exposes a
# Unix socket for Bazel.
#
# Bazel's remote cache client cannot use an HTTP proxy (only a unix: socket, see
# CombinedCacheClientFactory), so on a proxy-only network it cannot reach GCS
# directly. rbe-proxy-server bridges the gap: Bazel connects over the Unix socket,
# and the proxy reaches the GCS bucket through HTTPS_PROXY (Go's transport honors
# the proxy env), attaching a bearer token from Application Default Credentials.
# A local disk cache in ~/.cache/bazel/remote absorbs repeated reads and keeps
# the bucket in the raw /cas, /ac layout so it stays usable directly too.
#
# Requires: Application Default Credentials with access to the bucket. Use a
# service account (GOOGLE_APPLICATION_CREDENTIALS) rather than user credentials,
# which reauth policies can reject non-interactively (invalid_rapt).
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../../.."

remote_cache="https://ndscm-prod-rbe-cache-prod.storage.googleapis.com/"
cache_dir="${HOME}/.cache/bazel/remote"
socket="${cache_dir}/proxy-${RANDOM}.sock"

mkdir -p "${cache_dir}"

# Point Bazel at the proxy while it runs. rbe.local.bazelrc is per-developer
# (gitignored) and loaded via `try-import` in .bazelrc. Remove it (and the
# socket) on exit so a stopped proxy doesn't leave Bazel pointed at a dead
# socket.
#
# --remote_cache only turns the HTTP cache on and supplies the request scheme
# (must be http, since the Unix socket speaks plaintext); its host is never
# dialed because --remote_proxy routes the connection to the socket, so the host
# is a placeholder.
#
# --google_default_credentials=false overrides the workspace default: the proxy
# socket is local and unauthenticated, and the proxy does the GCS auth itself, so
# Bazel must not fetch Application Default Credentials (which hangs on the GCE
# metadata server off-GCP -> "Error initializing RemoteModule").
cat >rbe.local.bazelrc <<EOF
common --remote_cache=http://placeholder
common --google_default_credentials=false
common --remote_proxy=unix:${socket}
EOF
trap 'rm -f rbe.local.bazelrc "${socket}"' EXIT

# Build the proxy and run the binary directly. Not `bazel run`, which holds the
# workspace lock for its whole lifetime and would block the builds this proxy
# is meant to serve.
#
# --remote_cache= disables the cache for this bootstrap build: rbe.local.bazelrc
# has already pointed Bazel at this proxy's socket, which is not up yet.
bazel build --remote_cache= //seed/devprod/rbe/proxy/server:rbe-proxy-server
bazel-bin/seed/devprod/rbe/proxy/server/rbe-proxy-server_/rbe-proxy-server \
  --remote_cache="${remote_cache}" \
  --google_default_credentials=true \
  --local_cache="${cache_dir}" \
  --listen="unix:${socket}"
