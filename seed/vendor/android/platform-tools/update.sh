#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

# See: https://developer.android.com/tools/releases/platform-tools
tag="${1:-"r37.0.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/android/platform-tools/adb.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/android/platform-tools/fastboot.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/android/platform-tools/bin"

chmod +x ./seed/vendor/android/platform-tools/bin/adb.dotslash
chmod +x ./seed/vendor/android/platform-tools/bin/fastboot.dotslash

ln -s -f adb.dotslash ./seed/vendor/android/platform-tools/bin/adb
ln -s -f fastboot.dotslash ./seed/vendor/android/platform-tools/bin/fastboot
