#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

printf "\e[34mChecking golang tools...\e[0m\n"

if [[ "${oslike}" == "debian" ]]; then
  if [[ "${run_sudo}" == "true" ]]; then
    sudo -E snap install --classic go
  else
    printf "\e[31mSkipping snap go installation\e[0m\n"
  fi
fi

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  cat <<EOF >>"${HOME}/.seed_managed_profile.tmp"

## Golang

export PATH="\${HOME}/go/bin:\${PATH}"
EOF

  if command -v go >/dev/null 2>&1; then
    go install -v github.com/bazelbuild/buildtools/buildifier@latest
    go install -v github.com/bazelbuild/buildtools/buildozer@latest
    go install -v github.com/cweill/gotests/gotests@latest
    go install -v github.com/fatih/gomodifytags@latest
    go install -v github.com/go-delve/delve/cmd/dlv@latest
    go install -v github.com/haya14busa/goplay/cmd/goplay@latest
    go install -v github.com/josharian/impl@latest
    go install -v golang.org/x/tools/cmd/goimports@latest
    go install -v golang.org/x/tools/gopls@latest
    go install -v honnef.co/go/tools/cmd/staticcheck@latest
  else
    printf "\e[31mgo not found on PATH, skipping go tool installation\e[0m\n"
  fi
fi

printf "\e[32mCheck golang tools done.\e[0m\n"
