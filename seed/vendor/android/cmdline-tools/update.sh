#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

# See: https://developer.android.com/studio#command-line-tools-only
tag="${1:-"15859902"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/android/cmdline-tools/android.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/android/cmdline-tools/apkanalyzer.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/android/cmdline-tools/avdmanager.dotslash.json" \
  --skeleton "$(pwd)/seed/vendor/android/cmdline-tools/sdkmanager.dotslash.json" \
  --replace "TAG=${tag}" \
  --outdir "$(pwd)/seed/vendor/android/cmdline-tools/bin"

chmod +x ./seed/vendor/android/cmdline-tools/bin/android.dotslash
chmod +x ./seed/vendor/android/cmdline-tools/bin/apkanalyzer.dotslash
chmod +x ./seed/vendor/android/cmdline-tools/bin/avdmanager.dotslash
chmod +x ./seed/vendor/android/cmdline-tools/bin/sdkmanager.dotslash

ln -s -f android.dotslash ./seed/vendor/android/cmdline-tools/bin/android
ln -s -f apkanalyzer.dotslash ./seed/vendor/android/cmdline-tools/bin/apkanalyzer
ln -s -f avdmanager.dotslash ./seed/vendor/android/cmdline-tools/bin/avdmanager
ln -s -f sdkmanager.dotslash ./seed/vendor/android/cmdline-tools/bin/sdkmanager
