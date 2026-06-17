#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

# seed/devprod/smd/workstation/maintain-01-os.sh

uname="$(uname)"
distro="unknown"
oslike="unknown"
wsl="false"
run_sudo="true"

if [[ "${uname}" == "Linux" ]]; then
  if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "${ID}" == "ubuntu" ]]; then
      distro="ubuntu"
      oslike="debian"
    fi
  fi
elif [[ "${uname}" == "Darwin" ]]; then
  distro="darwin"
  oslike="darwin"
fi

if [[ "${uname}" == "Linux" && "${distro}" == "ubuntu" && "${oslike}" == "debian" ]]; then
  printf "\e[33mSystem Distro: ${distro}\nPackage Manager: ${oslike}\e[0m\n"
elif [[ "${uname}" == "Darwin" && "${distro}" == "darwin" && "${oslike}" == "darwin" ]]; then
  printf "\e[33mSystem Distro: ${distro}\nPackage Manager: ${oslike}\e[0m\n"
else
  printf "\e[31mUnsupported system: ${distro} (${oslike})\e[0m\n"
  exit 1
fi

if [[ ! -z "${WSL_DISTRO_NAME+x}" ]]; then
  wsl="true"
  printf "\e[33mWSL: true\e[0m\n"
fi

if [[ "${oslike}" == "darwin" ]]; then
  if ! xcode-select -p >/dev/null 2>&1; then
    printf "\e[33mXcode Command Line Tools not found. Installing...\e[0m\n"
    xcode-select --install
  fi
  if ! xcodebuild -license check >/dev/null 2>&1; then
    printf "\e[33mXcode license is NOT accepted yet...\e[0m\n"
    if [[ "${run_sudo}" == "true" ]]; then
      sudo xcodebuild -license accept
    else
      printf "\e[31mPlease accept the license with\n    sudo xcodebuild -license accept\e[0m\n"
      exit 1
    fi
  fi
fi

# seed/devprod/smd/workstation/maintain-02-identity.sh

if [[ -z "${ND_USER_HANDLE+x}" ]]; then
  read -p "Enter user handle (username before @company.com): " ND_USER_HANDLE
fi
printf "\e[33mUser handle: ${ND_USER_HANDLE}\e[0m\n"

if [[ -z "${ND_USER_DOMAIN+x}" ]]; then
  read -p "Enter user domain (domain after ${ND_USER_HANDLE}@): " ND_USER_DOMAIN
fi
printf "\e[33mUser email: ${ND_USER_HANDLE}@${ND_USER_DOMAIN}\e[0m\n"

if [[ -z "${ND_USER_DISPLAY_NAME+x}" ]]; then
  read -p "Enter user display name: " ND_USER_DISPLAY_NAME
fi
printf "\e[33mUser display Name: ${ND_USER_DISPLAY_NAME}\e[0m\n"

export ND_USER_HANDLE
export ND_USER_DOMAIN
export ND_USER_EMAIL="${ND_USER_HANDLE}@${ND_USER_DOMAIN}"
export ND_USER_DISPLAY_NAME

# seed/devprod/smd/workstation/maintain-11-basic-packages.sh

printf "\e[34mChecking basic packages...\e[0m\n"

if [[ "${oslike}" == "debian" ]]; then
  if [[ "${run_sudo}" == "true" ]]; then
    sudo apt update
    sudo apt upgrade -y
    sudo apt install -y curl
    sudo apt install -y direnv
    sudo apt install -y git
    sudo apt install -y netcat-openbsd
    sudo apt install -y ssh
  else
    printf "\e[31mSkipping system package installation\e[0m\n"
  fi
fi

if [[ "${oslike}" == "darwin" ]]; then
  if [[ "${run_sudo}" == "true" ]]; then
    if brew --version; then
      printf "\e[33mFound brew, skip install homebrew.\e[0m\n"
    else
      /bin/bash -c "$(curl -fsSL https://raw.githubusercontent.com/Homebrew/install/HEAD/install.sh)"
    fi
    if ! grep -F -q -s -x 'eval "$(/opt/homebrew/bin/brew shellenv)"' "${HOME}/.zprofile"; then
      printf '\neval "$(/opt/homebrew/bin/brew shellenv)"\n' >>"${HOME}/.zprofile"
    fi
    eval "$(/opt/homebrew/bin/brew shellenv)"

    brew install direnv
    brew install socat
  else
    printf "\e[31mSkipping homebrew and package installation\e[0m\n"
  fi
fi

printf "\e[32mCheck basic packages done.\e[0m\n"

# seed/devprod/smd/workstation/maintain-12-ssh-identity.sh

printf "\e[34mChecking ssh key pair...\e[0m\n"

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  if [[ -f "${HOME}/.ssh/id_ed25519" ]]; then
    printf "\e[33mFound .ssh/ed25519, skip regeneration.\e[0m\n"
  else
    printf "\e[34mGenerating ssh key pair for ${ND_USER_HANDLE}@${ND_USER_DOMAIN} with ed25519 algorithm.\e[0m\n"
    read -p "Enter workstation tag (e.g. t14wsl): " workstation_tag
    ssh-keygen -t ed25519 -C "${ND_USER_HANDLE}+${workstation_tag}@${ND_USER_DOMAIN}"
    public_key=$(cat "${HOME}/.ssh/id_ed25519.pub")
    printf "\e[33mCopy your public key to your github account:\n    ${public_key}\e[0m\n"
    read -p "Press <Enter> to continue..."
  fi
fi

printf "\e[32mCheck ssh key pair done.\e[0m\n"

# seed/devprod/smd/workstation/maintain-32-managed-proxy.sh

printf "\e[34mChecking managed proxy...\e[0m\n"

if [[ -n "${HTTPS_PROXY:-}" && -n "${NO_PROXY:-}" ]]; then
  proxy_host=$(echo "${HTTPS_PROXY}" | sed -E 's#^https?://([^:/]+)(:[0-9]+)?/?#\1#')
  proxy_port=$(echo "${HTTPS_PROXY}" | sed -E 's#^https?://[^:/]+(:[0-9]+)?/?#\1#' | sed -E 's#^:([0-9]+)$#\1#')
  if [[ -z "${proxy_port}" ]]; then
    proxy_port="443"
  fi

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    if [[ -f "${HOME}/.ssh/config" && -n "$(sed -n '/Host github.com/{N;/ProxyCommand/p;}' "${HOME}/.ssh/config")" ]]; then
      printf "\e[33mFound ssh proxy to github.com for git, skip.\e[0m\n"
    else
      printf "\e[34mAdding ssh proxy to github.com for git...\e[0m\n"
      if [[ "${oslike}" == "debian" ]]; then
        printf "\n%s\n%s\n" \
          "Host github.com" \
          "    ProxyCommand nc -X connect -x ${proxy_host}:${proxy_port} ssh.github.com 443" \
          >>"${HOME}/.ssh/config"
      fi
      if [[ "${oslike}" == "darwin" ]]; then
        printf "\n%s\n%s\n" \
          "Host github.com" \
          "    ProxyCommand socat - \"PROXY:${proxy_host}:ssh.github.com:443,proxyport=${proxy_port}\"" \
          >>"${HOME}/.ssh/config"
      fi
    fi
  fi
fi

printf "\e[32mCheck managed proxy done.\e[0m\n"

# seed/devprod/smd/workstation/maintain-51-monorepo.sh

printf "\e[34mChecking theseed monorepo...\e[0m\n"

if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
  mkdir -p "${HOME}/theseed"

  if [[ -d "${HOME}/theseed/theseed.git" && -d "${HOME}/theseed/main" ]]; then
    printf "\e[33mFound existing theseed monorepo, skip clone.\e[0m\n"
  else
    printf "\e[34mCloning theseed monorepo...\e[0m\n"
    git clone --bare --single-branch \
      --config "core.logallrefupdates=true" \
      --config "remote.origin.fetch=+refs/heads/*:refs/remotes/origin/*" \
      git@github.com:ndscm/theseed.git \
      "${HOME}/theseed/theseed.git"
    git --git-dir "${HOME}/theseed/theseed.git" worktree add -B main "${HOME}/theseed/main" origin/main
  fi
  git --git-dir "${HOME}/theseed/theseed.git" config user.name "${ND_USER_DISPLAY_NAME}"
  git --git-dir "${HOME}/theseed/theseed.git" config user.email "${ND_USER_HANDLE}@${ND_USER_DOMAIN}"

  cd "${HOME}/theseed/main"
  git fetch --all --prune
  git rebase origin/main
fi

printf "\e[34mCheck theseed monorepo done.\e[0m\n"

# seed/devprod/smd/workstation/maintain.sh

exec "${HOME}/theseed/main/seed/devprod/smd/workstation/maintain.sh"
