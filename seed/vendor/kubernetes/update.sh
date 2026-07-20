#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../.."

kubectl="${1:-"v1.36.2"}"
minikube="${2:-"v1.38.1"}"

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/kubernetes/kubectl.dotslash.json" \
  --replace "TAG=${kubectl}" \
  >./seed/vendor/kubernetes/bin/kubectl
chmod +x ./seed/vendor/kubernetes/bin/kubectl

bazel run //seed/devprod/dotslash/update -- \
  --skeleton "$(pwd)/seed/vendor/kubernetes/minikube.dotslash.json" \
  --replace "TAG=${minikube}" \
  >./seed/vendor/kubernetes/bin/minikube
chmod +x ./seed/vendor/kubernetes/bin/minikube
