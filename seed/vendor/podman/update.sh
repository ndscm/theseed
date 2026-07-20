#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

# Podman ships no static full engine binary; only the podman-remote client is
# available as a static/prebuilt binary. Tag is the version without the leading
# "v" so {{TAG}} also matches the versioned dir inside the macOS/Windows zips.
tag="${1:-"5.8.3"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/podman/podman-remote.dotslash.json" \
  --replace "TAG=${tag}" \
  >./seed/vendor/podman/bin/podman-remote
chmod +x ./seed/vendor/podman/bin/podman-remote
