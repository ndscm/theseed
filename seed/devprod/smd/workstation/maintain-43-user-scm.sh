#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking scm configs...\e[0m\n"

  git config --global core.autocrlf input
  git config --global diff.tool p4merge
  if ${wsl}; then
    git config --global difftool.p4merge.cmd 'p4merge.exe "$LOCAL" "$REMOTE"'
  else
    git config --global difftool.p4merge.cmd 'p4merge "$LOCAL" "$REMOTE"'
  fi
  git config --global difftool.p4merge.trustExitCode "true"
  git config --global merge.tool p4merge
  if ${wsl}; then
    git config --global mergetool.p4merge.cmd 'p4merge.exe "$BASE" "$LOCAL" "$REMOTE" "$MERGED"'
  else
    git config --global mergetool.p4merge.cmd 'p4merge "$BASE" "$LOCAL" "$REMOTE" "$MERGED"'
  fi
  git config --global mergetool.p4merge.trustExitCode "true"
  git config --global mergetool.keepBackup "false"

  printf "\e[32m[user] Check scm configs done.\e[0m\n"
fi
