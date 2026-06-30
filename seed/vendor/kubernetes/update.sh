#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

tag="v1.36.2"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/kubernetes/kubectl.dotslash.json" \
  --no-format \
  --tag="${tag}" \
  >./seed/vendor/kubernetes/bin/kubectl
chmod +x ./seed/vendor/kubernetes/bin/kubectl

minikube="v1.38.1"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/kubernetes/minikube.dotslash.json" \
  --no-format \
  --tag="${minikube}" \
  >./seed/vendor/kubernetes/bin/minikube
chmod +x ./seed/vendor/kubernetes/bin/minikube
