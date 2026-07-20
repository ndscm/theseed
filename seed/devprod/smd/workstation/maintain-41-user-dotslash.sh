#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking dotslash tools...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    cp -a -f ./seed/vendor/bazel/bin/bazel "${HOME}/.local/bin/bazel"
    cp -a -f ./seed/vendor/bazel/bin/bazelisk "${HOME}/.local/bin/bazelisk"
    cp -a -f ./seed/vendor/bazel/bin/bazelisk.dotslash "${HOME}/.local/bin/bazelisk.dotslash"
    cp -a -f ./seed/vendor/buck/bin/buck2 "${HOME}/.local/bin/buck2"
    cp -a -f ./seed/vendor/buck/bin/buck2.dotslash "${HOME}/.local/bin/buck2.dotslash"
    cp -a -f ./seed/vendor/buck/bin/reindeer "${HOME}/.local/bin/reindeer"
    cp -a -f ./seed/vendor/buck/bin/reindeer.dotslash "${HOME}/.local/bin/reindeer.dotslash"
    cp -a -f ./seed/vendor/claude/bin/claude "${HOME}/.local/bin/claude"
    cp -a -f ./seed/vendor/claude/bin/claude.dotslash "${HOME}/.local/bin/claude.dotslash"
    cp -a -f ./seed/vendor/crane/bin/crane "${HOME}/.local/bin/crane"
    cp -a -f ./seed/vendor/crane/bin/crane.dotslash "${HOME}/.local/bin/crane.dotslash"
    cp -a -f ./seed/vendor/github/bin/gh "${HOME}/.local/bin/gh"
    cp -a -f ./seed/vendor/github/bin/gh.dotslash "${HOME}/.local/bin/gh.dotslash"
    cp -a -f ./seed/vendor/jq/bin/jq "${HOME}/.local/bin/jq"
    cp -a -f ./seed/vendor/jq/bin/jq.dotslash "${HOME}/.local/bin/jq.dotslash"
    cp -a -f ./seed/vendor/lego/bin/lego "${HOME}/.local/bin/lego"
    cp -a -f ./seed/vendor/lego/bin/lego.dotslash "${HOME}/.local/bin/lego.dotslash"
    cp -a -f ./seed/vendor/llvm/bin/clang "${HOME}/.local/bin/clang"
    cp -a -f ./seed/vendor/llvm/bin/clang-format "${HOME}/.local/bin/clang-format"
    cp -a -f ./seed/vendor/llvm/bin/clang-format.dotslash "${HOME}/.local/bin/clang-format.dotslash"
    cp -a -f ./seed/vendor/llvm/bin/clang.dotslash "${HOME}/.local/bin/clang.dotslash"
    cp -a -f ./seed/vendor/llvm/bin/clangd "${HOME}/.local/bin/clangd"
    cp -a -f ./seed/vendor/llvm/bin/clangd.dotslash "${HOME}/.local/bin/clangd.dotslash"
    cp -a -f ./seed/vendor/ndscm/bin/ndscm "${HOME}/.local/bin/ndscm"
    cp -a -f ./seed/vendor/ndscm/bin/ndscm.dotslash "${HOME}/.local/bin/ndscm.dotslash"
    cp -a -f ./seed/vendor/node/bin/corepack "${HOME}/.local/bin/corepack"
    cp -a -f ./seed/vendor/node/bin/corepack.dotslash "${HOME}/.local/bin/corepack.dotslash"
    cp -a -f ./seed/vendor/node/bin/node "${HOME}/.local/bin/node"
    cp -a -f ./seed/vendor/node/bin/node.dotslash "${HOME}/.local/bin/node.dotslash"
    cp -a -f ./seed/vendor/node/bin/npm "${HOME}/.local/bin/npm"
    cp -a -f ./seed/vendor/node/bin/npm.dotslash "${HOME}/.local/bin/npm.dotslash"
    cp -a -f ./seed/vendor/node/bin/npx "${HOME}/.local/bin/npx"
    cp -a -f ./seed/vendor/node/bin/npx.dotslash "${HOME}/.local/bin/npx.dotslash"
    cp -a -f ./seed/vendor/node/bin/pnpm "${HOME}/.local/bin/pnpm"
    cp -a -f ./seed/vendor/node/bin/pnpm.shim "${HOME}/.local/bin/pnpm.shim"
    cp -a -f ./seed/vendor/node/bin/pnpx "${HOME}/.local/bin/pnpx"
    cp -a -f ./seed/vendor/node/bin/pnpx.shim "${HOME}/.local/bin/pnpx.shim"
    cp -a -f ./seed/vendor/node/bin/yarn "${HOME}/.local/bin/yarn"
    cp -a -f ./seed/vendor/node/bin/yarn.shim "${HOME}/.local/bin/yarn.shim"
    cp -a -f ./seed/vendor/node/bin/yarnpkg "${HOME}/.local/bin/yarnpkg"
    cp -a -f ./seed/vendor/node/bin/yarnpkg.shim "${HOME}/.local/bin/yarnpkg.shim"
    cp -a -f ./seed/vendor/podman/bin/podman-remote "${HOME}/.local/bin/podman-remote"
    cp -a -f ./seed/vendor/podman/bin/podman-remote.dotslash "${HOME}/.local/bin/podman-remote.dotslash"
    cp -a -f ./seed/vendor/rust/bin/rustup "${HOME}/.local/bin/rustup"
    cp -a -f ./seed/vendor/rust/bin/rustup-init "${HOME}/.local/bin/rustup-init"
    cp -a -f ./seed/vendor/rust/bin/rustup-init.dotslash "${HOME}/.local/bin/rustup-init.dotslash"
    cp -a -f ./seed/vendor/uv/bin/uv "${HOME}/.local/bin/uv"
    cp -a -f ./seed/vendor/uv/bin/uv.dotslash "${HOME}/.local/bin/uv.dotslash"
    cp -a -f ./seed/vendor/uv/bin/uvx "${HOME}/.local/bin/uvx"
    cp -a -f ./seed/vendor/uv/bin/uvx.dotslash "${HOME}/.local/bin/uvx.dotslash"
  fi

  if [[ "${oslike}" == "debian" ]]; then
    cp -a -f ./contrib/vendor/perforce/bin/p4merge "${HOME}/.local/bin/p4merge"
    cp -a -f ./contrib/vendor/perforce/bin/p4merge.dotslash "${HOME}/.local/bin/p4merge.dotslash"
  fi

  printf "\e[32m[user] Check dotslash tools done.\e[0m\n"
fi
