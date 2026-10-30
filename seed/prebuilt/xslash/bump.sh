#!/usr/bin/env bash
set -euo pipefail
cd "$(dirname "${BASH_SOURCE[0]}")"

tag="${1:-"$(cat ./TAG)"}"

if [[ "${1:-}" == "latest" ]]; then
  latest="$(curl -s "https://api.github.com/repos/ndscm/xslash/releases/latest")"
  tag="$(jq -r ".tag_name" <<<"${latest}")"
  printf "%s\n" "${tag}" >./TAG
fi

release="$(curl -s "https://api.github.com/repos/ndscm/xslash/releases/tags/${tag}")"

platforms=(
  "darwin-amd64"
  "darwin-arm64"
  "linux-amd64"
  "linux-arm64"
)

{
  printf '%s\n' 'http_file = use_repo_rule("@bazel_tools//tools/build_defs/repo:http.bzl", "http_file")'
  for platform in "${platforms[@]}"; do
    url="https://github.com/ndscm/xslash/releases/download/${tag}/xslash-${platform}"
    digest="$(jq -r --arg name "xslash-${platform}" '.assets[] | select(.name == $name) | .digest' <<<"${release}")"
    integrity="sha256-$(printf '%s' "${digest#sha256:}" | xxd -r -p | openssl base64 -A)"
    printf '\n'
    printf 'http_file(\n'
    printf '    name = "xslash-%s",\n' "${platform}"
    printf '    executable = True,\n'
    printf '    integrity = "%s",\n' "${integrity}"
    printf '    urls = ["%s"],\n' "${url}"
    printf ')\n'
  done
  for platform in "${platforms[@]}"; do
    url="https://github.com/ndscm/xslash/releases/download/${tag}/xslash-helper-${platform}"
    digest="$(jq -r --arg name "xslash-helper-${platform}" '.assets[] | select(.name == $name) | .digest' <<<"${release}")"
    integrity="sha256-$(printf '%s' "${digest#sha256:}" | xxd -r -p | openssl base64 -A)"
    printf '\n'
    printf 'http_file(\n'
    printf '    name = "xslash-helper-%s",\n' "${platform}"
    printf '    executable = True,\n'
    printf '    integrity = "%s",\n' "${integrity}"
    printf '    urls = ["%s"],\n' "${url}"
    printf ')\n'
  done
} >./xslash.MODULE.bazel
