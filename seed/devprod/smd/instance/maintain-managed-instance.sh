#!/usr/bin/env bash
set -eux
set -o pipefail

# # Detect distro

uname="$(uname)"
distro="unknown"
oslike="unknown"

if [[ "${uname}" == "Darwin" ]]; then
  distro="darwin"
  oslike="darwin"
fi
if [[ "${uname}" == "Linux" ]]; then
  if [[ -f /etc/os-release ]]; then
    . /etc/os-release
    if [[ "${ID}" == "debian" ]]; then
      distro="debian"
      oslike="debian"
    fi
    if [[ "${ID}" == "fedora" ]]; then
      distro="fedora"
      oslike="fedora"
    fi
    if [[ "${ID}" == "rocky" ]]; then
      distro="rocky"
      oslike="fedora"
    fi
    if [[ "${ID}" == "ubuntu" ]]; then
      distro="ubuntu"
      oslike="debian"
    fi
  fi
fi

if [[ "${uname}" == "Linux" && "${distro}" == "fedora" && "${oslike}" == "fedora" ]]; then
  : # support fedora
elif [[ "${uname}" == "Linux" && "${distro}" == "rocky" && "${oslike}" == "fedora" ]]; then
  : # support rocky
else
  printf "\e[31mUnsupported system: ${distro} (${oslike})\e[0m\n"
  exit 1
fi
printf "\e[33mSystem Distro: ${distro}\nPackage Manager: ${oslike}\e[0m\n"

# # Configure system

# ## Swapfile

if [[ "${oslike}" == "fedora" ]]; then
  if [[ ! -f "/swapfile" ]]; then
    sudo dd if=/dev/zero of=/swapfile bs=1MB count=4096
    sudo chmod 0600 /swapfile
    sudo mkswap /swapfile
    sudo swapon /swapfile
    cat <<EOF | sudo tee --append /etc/fstab
/swapfile swap swap defaults 0 0
EOF
  fi
fi

# ## Over commit

if [[ "${oslike}" == "fedora" ]]; then
  if [[ ! -f "/usr/lib/sysctl.d/90-theseed-redis.conf" ]]; then
    cat <<EOF | sudo tee --append /usr/lib/sysctl.d/90-theseed-redis.conf
vm.overcommit_memory = 1
EOF
  fi
fi

# # Install system dependencies

if [[ "${oslike}" == "fedora" ]]; then
  sudo dnf update -y
fi

# # Install cockpit

if [[ "${oslike}" == "fedora" ]]; then
  sudo systemctl enable --now cockpit.socket
fi

# # Install podman

if [[ "${oslike}" == "fedora" ]]; then
  sudo dnf install -y podman
  if [[ ! "$(getent group podman)" ]]; then
    sudo groupadd --system podman
  fi
  sudo mkdir -p /etc/systemd/system/podman.socket.d
  cat <<EOF | sudo tee /etc/systemd/system/podman.socket.d/override.conf
[Socket]
SocketUser=root
SocketGroup=podman
EOF
  sudo systemctl daemon-reload
  sudo systemctl enable --now podman.socket
  sudo chown root:podman /run/podman
  sudo chmod 0750 /run/podman
fi

sudo usermod -a -G podman $(whoami)
