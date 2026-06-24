#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking dotslash tools...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    cp -f ./seed/vendor/bazel/bin/bazel "${HOME}/.local/bin/bazel"
    cp -f ./seed/vendor/bazel/bin/bazelisk "${HOME}/.local/bin/bazelisk"
    cp -f ./seed/vendor/buck/bin/buck2 "${HOME}/.local/bin/buck2"
    cp -f ./seed/vendor/buck/bin/reindeer "${HOME}/.local/bin/reindeer"
    cp -f ./seed/vendor/claude/bin/claude "${HOME}/.local/bin/claude"
    cp -f ./seed/vendor/crane/bin/crane "${HOME}/.local/bin/crane"
    cp -f ./seed/vendor/github/bin/gh "${HOME}/.local/bin/gh"
    cp -f ./seed/vendor/jq/bin/jq "${HOME}/.local/bin/jq"
    cp -f ./seed/vendor/lego/bin/lego "${HOME}/.local/bin/lego"
    cp -f ./seed/vendor/llvm/bin/clang "${HOME}/.local/bin/clang"
    cp -f ./seed/vendor/llvm/bin/clang-format "${HOME}/.local/bin/clang-format"
    cp -f ./seed/vendor/llvm/bin/clangd "${HOME}/.local/bin/clangd"
    cp -f ./seed/vendor/ndscm/bin/ndscm "${HOME}/.local/bin/ndscm"
    cp -f ./seed/vendor/node/bin/corepack "${HOME}/.local/bin/corepack"
    cp -f ./seed/vendor/node/bin/node "${HOME}/.local/bin/node"
    cp -f ./seed/vendor/node/bin/npm "${HOME}/.local/bin/npm"
    cp -f ./seed/vendor/node/bin/npx "${HOME}/.local/bin/npx"
    cp -f ./seed/vendor/node/bin/pnpm "${HOME}/.local/bin/pnpm"
    cp -f ./seed/vendor/node/bin/pnpx "${HOME}/.local/bin/pnpx"
    cp -f ./seed/vendor/node/bin/yarn "${HOME}/.local/bin/yarn"
    cp -f ./seed/vendor/node/bin/yarnpkg "${HOME}/.local/bin/yarnpkg"
    cp -f ./seed/vendor/podman/bin/podman-remote "${HOME}/.local/bin/podman-remote"
    cp -f ./seed/vendor/rust/bin/rustup "${HOME}/.local/bin/rustup"
    cp -f ./seed/vendor/rust/bin/rustup-init "${HOME}/.local/bin/rustup-init"
    cp -f ./seed/vendor/uv/bin/uv "${HOME}/.local/bin/uv"
    cp -f ./seed/vendor/uv/bin/uvx "${HOME}/.local/bin/uvx"
  fi

  if [[ "${oslike}" == "debian" ]]; then
    cp -f ./contrib/vendor/perforce/bin/p4merge "${HOME}/.local/bin/p4merge"
  fi

  printf "\e[32m[user] Check dotslash tools done.\e[0m\n"
fi
