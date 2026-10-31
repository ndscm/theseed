#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking dotslash tools...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    xslash -- create --skeleton ./seed/prebuilt/age/bin/age-inspect.xslash --out "${HOME}/.local/bin/age-inspect.xslash" --symlink age-inspect
    xslash -- create --skeleton ./seed/prebuilt/age/bin/age-keygen.xslash --out "${HOME}/.local/bin/age-keygen.xslash" --symlink age-keygen
    xslash -- create --skeleton ./seed/prebuilt/age/bin/age.xslash --out "${HOME}/.local/bin/age.xslash" --symlink age
    xslash -- create --skeleton ./seed/prebuilt/github/bin/gh.xslash --out "${HOME}/.local/bin/gh.xslash" --symlink gh
    xslash -- create --skeleton ./seed/prebuilt/go/bin/go.xslash --out "${HOME}/.local/bin/go.xslash" --symlink go
    xslash -- create --skeleton ./seed/prebuilt/go/bin/gofmt.xslash --out "${HOME}/.local/bin/gofmt.xslash" --symlink gofmt
    xslash -- create --skeleton ./seed/prebuilt/ndscm/bin/ndscm.xslash --out "${HOME}/.local/bin/ndscm.xslash" --symlink nd --symlink ndscm
    xslash -- create --skeleton ./seed/prebuilt/node/bin/node.xslash --out "${HOME}/.local/bin/node.xslash" --symlink node
    xslash -- create --skeleton ./seed/prebuilt/node/bin/npm.xslash --out "${HOME}/.local/bin/npm.xslash" --symlink npm
    xslash -- create --skeleton ./seed/prebuilt/node/bin/npx.xslash --out "${HOME}/.local/bin/npx.xslash" --symlink npx
    xslash -- create --skeleton ./seed/prebuilt/pnpm/bin/pnpm.xslash --out "${HOME}/.local/bin/pnpm.xslash" --symlink pnpm
    xslash -- create --skeleton ./seed/prebuilt/python/bin/python3.xslash --out "${HOME}/.local/bin/python3.xslash" --symlink python --symlink python3
    xslash -- create --skeleton ./seed/prebuilt/rust/bin/cargo-clippy.xslash --out "${HOME}/.local/bin/cargo-clippy.xslash" --symlink cargo-clippy
    xslash -- create --skeleton ./seed/prebuilt/rust/bin/cargo-fmt.xslash --out "${HOME}/.local/bin/cargo-fmt.xslash" --symlink cargo-fmt
    xslash -- create --skeleton ./seed/prebuilt/rust/bin/cargo.xslash --out "${HOME}/.local/bin/cargo.xslash" --symlink cargo
    xslash -- create --skeleton ./seed/prebuilt/rust/bin/rust-analyzer.xslash --out "${HOME}/.local/bin/rust-analyzer.xslash" --symlink rust-analyzer
    xslash -- create --skeleton ./seed/prebuilt/rust/bin/rust-gdb.xslash --out "${HOME}/.local/bin/rust-gdb.xslash" --symlink rust-gdb
    xslash -- create --skeleton ./seed/prebuilt/rust/bin/rustc.xslash --out "${HOME}/.local/bin/rustc.xslash" --symlink rustc
    xslash -- create --skeleton ./seed/prebuilt/rust/bin/rustdoc.xslash --out "${HOME}/.local/bin/rustdoc.xslash" --symlink rustdoc
    xslash -- create --skeleton ./seed/prebuilt/rust/bin/rustfmt.xslash --out "${HOME}/.local/bin/rustfmt.xslash" --symlink rustfmt
    xslash -- create --skeleton ./seed/prebuilt/rustup/bin/rustup-init.xslash --out "${HOME}/.local/bin/rustup-init.xslash" --symlink rustup --symlink rustup-init
    xslash -- create --skeleton ./seed/prebuilt/shfmt/bin/shfmt.xslash --out "${HOME}/.local/bin/shfmt.xslash" --symlink shfmt
    xslash -- create --skeleton ./seed/prebuilt/uv/bin/uv.xslash --out "${HOME}/.local/bin/uv.xslash" --symlink uv
    xslash -- create --skeleton ./seed/prebuilt/uv/bin/uvx.xslash --out "${HOME}/.local/bin/uvx.xslash" --symlink uvx

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
    cp -a -f ./seed/vendor/podman/bin/podman-remote "${HOME}/.local/bin/podman-remote"
    cp -a -f ./seed/vendor/podman/bin/podman-remote.dotslash "${HOME}/.local/bin/podman-remote.dotslash"
    cp -a -f ./seed/vendor/sops/bin/sops "${HOME}/.local/bin/sops"
    cp -a -f ./seed/vendor/sops/bin/sops.dotslash "${HOME}/.local/bin/sops.dotslash"
  fi

  if [[ "${oslike}" == "debian" ]]; then
    cp -a -f ./contrib/vendor/perforce/bin/p4merge "${HOME}/.local/bin/p4merge"
    cp -a -f ./contrib/vendor/perforce/bin/p4merge.dotslash "${HOME}/.local/bin/p4merge.dotslash"
  fi

  printf "\e[32m[user] Check dotslash tools done.\e[0m\n"
fi
