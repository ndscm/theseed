#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking xslash tools...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    xslash -- create --skeleton ./seed/prebuilt/adb/bin/adb.xslash --out "${HOME}/.local/bin/adb.xslash" --symlink adb
    xslash -- create --skeleton ./seed/prebuilt/adb/bin/fastboot.xslash --out "${HOME}/.local/bin/fastboot.xslash" --symlink fastboot
    xslash -- create --skeleton ./seed/prebuilt/age/bin/age-inspect.xslash --out "${HOME}/.local/bin/age-inspect.xslash" --symlink age-inspect
    xslash -- create --skeleton ./seed/prebuilt/age/bin/age-keygen.xslash --out "${HOME}/.local/bin/age-keygen.xslash" --symlink age-keygen
    xslash -- create --skeleton ./seed/prebuilt/age/bin/age.xslash --out "${HOME}/.local/bin/age.xslash" --symlink age
    xslash -- create --skeleton ./seed/prebuilt/android/bin/android.xslash --out "${HOME}/.local/bin/android.xslash" --symlink android
    xslash -- create --skeleton ./seed/prebuilt/android/bin/apkanalyzer.xslash --out "${HOME}/.local/bin/apkanalyzer.xslash" --symlink apkanalyzer
    xslash -- create --skeleton ./seed/prebuilt/android/bin/avdmanager.xslash --out "${HOME}/.local/bin/avdmanager.xslash" --symlink avdmanager
    xslash -- create --skeleton ./seed/prebuilt/android/bin/sdkmanager.xslash --out "${HOME}/.local/bin/sdkmanager.xslash" --symlink sdkmanager
    xslash -- create --skeleton ./seed/prebuilt/atlas/bin/atlas.xslash --out "${HOME}/.local/bin/atlas.xslash" --symlink atlas
    xslash -- create --skeleton ./seed/prebuilt/bazel/bin/bazelisk.xslash --out "${HOME}/.local/bin/bazelisk.xslash" --symlink bazel --symlink bazelisk
    xslash -- create --skeleton ./seed/prebuilt/buck/bin/buck2.xslash --out "${HOME}/.local/bin/buck2.xslash" --symlink buck2
    xslash -- create --skeleton ./seed/prebuilt/claude/bin/claude.xslash --out "${HOME}/.local/bin/claude.xslash" --symlink claude
    xslash -- create --skeleton ./seed/prebuilt/crane/bin/crane.xslash --out "${HOME}/.local/bin/crane.xslash" --symlink crane
    xslash -- create --skeleton ./seed/prebuilt/gcloud/bin/gcloud.xslash --out "${HOME}/.local/bin/gcloud.xslash" --symlink gcloud
    xslash -- create --skeleton ./seed/prebuilt/github/bin/gh.xslash --out "${HOME}/.local/bin/gh.xslash" --symlink gh
    xslash -- create --skeleton ./seed/prebuilt/go/bin/go.xslash --out "${HOME}/.local/bin/go.xslash" --symlink go
    xslash -- create --skeleton ./seed/prebuilt/go/bin/gofmt.xslash --out "${HOME}/.local/bin/gofmt.xslash" --symlink gofmt
    xslash -- create --skeleton ./seed/prebuilt/jq/bin/jq.xslash --out "${HOME}/.local/bin/jq.xslash" --symlink jq
    xslash -- create --skeleton ./seed/prebuilt/kubectl/bin/kubectl.xslash --out "${HOME}/.local/bin/kubectl.xslash" --symlink kubectl
    xslash -- create --skeleton ./seed/prebuilt/lego/bin/lego.xslash --out "${HOME}/.local/bin/lego.xslash" --symlink lego
    xslash -- create --skeleton ./seed/prebuilt/llvm/bin/clang-format.xslash --out "${HOME}/.local/bin/clang-format.xslash" --symlink clang-format
    xslash -- create --skeleton ./seed/prebuilt/llvm/bin/clang.xslash --out "${HOME}/.local/bin/clang.xslash" --symlink clang
    xslash -- create --skeleton ./seed/prebuilt/llvm/bin/clangd.xslash --out "${HOME}/.local/bin/clangd.xslash" --symlink clangd
    xslash -- create --skeleton ./seed/prebuilt/minikube/bin/minikube.xslash --out "${HOME}/.local/bin/minikube.xslash" --symlink minikube
    xslash -- create --skeleton ./seed/prebuilt/ndscm/bin/ndscm.xslash --out "${HOME}/.local/bin/ndscm.xslash" --symlink nd --symlink ndscm
    xslash -- create --skeleton ./seed/prebuilt/node/bin/node.xslash --out "${HOME}/.local/bin/node.xslash" --symlink node
    xslash -- create --skeleton ./seed/prebuilt/node/bin/npm.xslash --out "${HOME}/.local/bin/npm.xslash" --symlink npm
    xslash -- create --skeleton ./seed/prebuilt/node/bin/npx.xslash --out "${HOME}/.local/bin/npx.xslash" --symlink npx
    xslash -- create --skeleton ./seed/prebuilt/pnpm/bin/pnpm.xslash --out "${HOME}/.local/bin/pnpm.xslash" --symlink pnpm
    xslash -- create --skeleton ./seed/prebuilt/podman/bin/podman-remote.xslash --out "${HOME}/.local/bin/podman-remote.xslash" --symlink podman-remote
    xslash -- create --skeleton ./seed/prebuilt/python/bin/python3.xslash --out "${HOME}/.local/bin/python3.xslash" --symlink python --symlink python3
    xslash -- create --skeleton ./seed/prebuilt/reindeer/bin/reindeer.xslash --out "${HOME}/.local/bin/reindeer.xslash" --symlink reindeer
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
    xslash -- create --skeleton ./seed/prebuilt/sops/bin/sops.xslash --out "${HOME}/.local/bin/sops.xslash" --symlink sops
    xslash -- create --skeleton ./seed/prebuilt/uv/bin/uv.xslash --out "${HOME}/.local/bin/uv.xslash" --symlink uv
    xslash -- create --skeleton ./seed/prebuilt/uv/bin/uvx.xslash --out "${HOME}/.local/bin/uvx.xslash" --symlink uvx

    xslash -- create --skeleton ./contrib/prebuilt/perforce/bin/p4merge.xslash --out "${HOME}/.local/bin/p4merge.xslash" --symlink p4merge
  fi

  printf "\e[32m[user] Check xslash tools done.\e[0m\n"
fi
