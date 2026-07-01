#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

maintain_scopes="${1:-""}"
export maintain_scopes

# seed/devprod/smd/workstation/maintain-00-os.sh

uname="$(uname)"
distro="unknown"
oslike="unknown"
wsl="false"

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

export uname
export distro
export oslike
export wsl

maintain_scopes="${maintain_scopes:-""}"

if [[ "${maintain_scopes}" == "" ]]; then
  maintain_scopes="user"
  if id -nG "$(id -un)" | tr ' ' '\n' | grep -qx -e sudo -e admin -e wheel; then
    read -p "You have sudo privileges. Use sudo for system maintenance? [y/N]: " use_sudo_reply
    if [[ "${use_sudo_reply,,}" == "y" || "${use_sudo_reply,,}" == "yes" ]]; then
      maintain_scopes="system,user"
    fi
  fi
fi

export maintain_scopes

# seed/devprod/smd/workstation/maintain-02-user-identity.sh

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
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
fi

# seed/devprod/smd/workstation/maintain-10-system-basic-packages.sh

if [[ ",${maintain_scopes}," == *",system,"* ]]; then
  printf "\e[34m[system] Checking basic packages...\e[0m\n"

  if [[ "${oslike}" == "debian" ]]; then
    sudo -E apt update
    sudo -E apt upgrade -y
    sudo -E apt install -y curl
    sudo -E apt install -y direnv
    sudo -E apt install -y git
    sudo -E apt install -y netcat-openbsd
    sudo -E apt install -y ssh
    sudo -E apt install -y tar
  fi

  if [[ "${oslike}" == "darwin" ]]; then
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
  fi

  printf "\e[32m[system] Check basic packages done.\e[0m\n"
fi

# seed/devprod/smd/workstation/maintain-11-user-basic-packages.sh

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking basic packages...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    mkdir -p "${HOME}/.local/bin"
    export PATH="${HOME}/.local/bin:${PATH}"
    curl -fsSL "https://github.com/ndscm/ndscm/releases/latest/download/ndscm.dotslash" >"${HOME}/.local/bin/ndscm"
    chmod +x "${HOME}/.local/bin/ndscm"
  fi

  if [[ "${oslike}" == "debian" ]]; then
    curl -fsSL "https://github.com/facebook/dotslash/releases/latest/download/dotslash-ubuntu-22.04.$(uname -m).tar.gz" |
      tar fxz - -C "${HOME}/.local/bin"
  fi

  if [[ "${oslike}" == "darwin" ]]; then
    curl -fsSL https://github.com/facebook/dotslash/releases/latest/download/dotslash-macos.tar.gz |
      tar fxz - -C "${HOME}/.local/bin"
  fi

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    curl -fsSL "https://raw.githubusercontent.com/ndscm/theseed/refs/heads/main/seed/vendor/direnv/bin/direnv" \
      >"${HOME}/.local/bin/direnv"
    chmod +x "${HOME}/.local/bin/direnv"
  fi

  printf "\e[32m[user] Check basic packages done.\e[0m\n"
fi

# seed/devprod/smd/workstation/maintain-12-user-ssh-identity.sh

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking ssh key pair...\e[0m\n"

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

  printf "\e[32m[user] Check ssh key pair done.\e[0m\n"
fi

# seed/devprod/smd/workstation/maintain-32-user-managed-proxy.sh

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking managed proxy...\e[0m\n"

  if [[ -n "${HTTPS_PROXY:-}" && -n "${NO_PROXY:-}" ]]; then
    proxy_host=$(printf "%s" "${HTTPS_PROXY}" | sed -E 's#^https?://([^:/]+)(:[0-9]+)?/?#\1#')
    proxy_port=$(printf "%s" "${HTTPS_PROXY}" | sed -E 's#^https?://[^:/]+(:[0-9]+)?/?#\1#' | sed -E 's#^:([0-9]+)$#\1#')
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

  printf "\e[32m[user] Check managed proxy done.\e[0m\n"
fi

# seed/devprod/smd/workstation/maintain-51-monorepo.sh

if [[ ",${maintain_scopes}," == *",user,"* ]]; then
  printf "\e[34m[user] Checking theseed monorepo...\e[0m\n"

  if [[ "${oslike}" == "debian" || "${oslike}" == "darwin" ]]; then
    mkdir -p "${HOME}/ndscm"

    if [[ -d "${HOME}/ndscm/theseed" ]]; then
      printf "\e[33mFound existing theseed monorepo, skip connect.\e[0m\n"
    else
      ndscm connect \
        --repos_home "${HOME}/ndscm" \
        theseed "git@github.com:ndscm/theseed.git"
    fi
  fi

  export ND_REPOS_HOME="${HOME}/ndscm"

  printf "\e[32m[user] Check theseed monorepo done.\e[0m\n"
fi

# seed/devprod/smd/workstation/maintain.sh

exec "${ND_REPOS_HOME}/theseed/main/seed/devprod/smd/workstation/maintain.sh" "${maintain_scopes}"
