#!/bin/bash
set -eux
set -o pipefail

if [[ -z "${ND_USER_HANDLE+x}" ]]; then
  read -p "Enter nd user handle (username before @ndscm.com): " ND_USER_HANDLE
fi
echo -e "\e[33mUsername: ${ND_USER_HANDLE}\e[0m"
export ND_USER_HANDLE
if [[ -z "${ND_USER_DISPLAY_NAME+x}" ]]; then
  read -p "Enter user display name: " ND_USER_DISPLAY_NAME
fi
echo -e "\e[33mDisplay Name: ${ND_USER_DISPLAY_NAME}\e[0m"
export ND_USER_DISPLAY_NAME

# # Prepare

distro="unknown"
oslike="unknown"

if [[ -f /etc/os-release ]]; then
  . /etc/os-release
  if [[ "${ID}" == "ubuntu" ]]; then
    distro="ubuntu"
    oslike="debian"
  fi
fi

if [[ "${distro}" == "ubuntu" && "${oslike}" == "debian" ]]; then
  echo -e "\e[33mSystem Distro: ${distro}\nPackage Manager: ${oslike}\e[0m"
else
  echo -e "\e[31mUnsupported system: ${distro} (${oslike})\e[0m"
  exit 1
fi

wsl=false
if [[ ! -z "${WSL_DISTRO_NAME+x}" ]]; then
  wsl=true
  printf "\e[33mWSL: true\e[0m\n"
fi








echo -e "\e[34mChecking basic packages...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  sudo apt update -y
  sudo apt upgrade -y
  sudo apt install -y git
  sudo apt install -y netcat-openbsd
  sudo apt install -y ssh
fi

echo -e "\e[32mCheck basic packages done.\e[0m"

# # SSH

echo -e "\e[34mChecking ssh key pair...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  if [[ -f "${HOME}/.ssh/id_ed25519" ]]; then
    echo -e "\e[33mFound .ssh/ed25519, skip regeneration.\e[0m"
  else
    echo -e "\e[34mGenerating ssh key pair for ${ND_USER_HANDLE}@ndscm.com with ed25519 algorithm.\e[0m"
    read -p "Enter workstation tag (e.g. t14wsl): " workstation_tag
    ssh-keygen -t ed25519 -C "${ND_USER_HANDLE}+${workstation_tag}@ndscm.com"
    public_key=$(cat ${HOME}/.ssh/id_ed25519.pub)
    echo -e "\e[33mCopy your public key to your github account:\n    ${public_key}\e[0m"
    read -p "Press <Enter> to continue..."
  fi
fi

echo -e "\e[32mCheck ssh key pair done.\e[0m"

# # Shell

echo -e "\e[34mChecking zsh shell...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  sudo apt install -y zsh
  if [[ ! -d ${HOME}/.oh-my-zsh ]]; then
    bash -c "$(curl -fsSL https://raw.githubusercontent.com/ohmyzsh/ohmyzsh/master/tools/install.sh)" -- --unattended
  fi
  cp ${HOME}/.oh-my-zsh/templates/zshrc.zsh-template ${HOME}/.zshrc
  sed -i 's/ZSH_THEME=".*"/ZSH_THEME="agnoster"/' ${HOME}/.zshrc
  echo 'unsetopt SHARE_HISTORY' >>${HOME}/.zshrc
fi

echo -e "\e[32mCheck zsh shell done.\e[0m"

echo -e "\e[34mChecking powerline fonts...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  if ! ${wsl}; then
    mkdir -p ${HOME}/github/powerline
    curl -o ${HOME}/github/powerline/fonts.tar.gz -L https://github.com/powerline/fonts/archive/refs/heads/master.tar.gz
    if [[ -d ${HOME}/github/powerline/fonts ]]; then
      rm -r ${HOME}/github/powerline/fonts
    fi
    mkdir -p ${HOME}/github/powerline/fonts
    tar -z -x -v --strip-components 1 -f ${HOME}/github/powerline/fonts.tar.gz -C ${HOME}/github/powerline/fonts/
    ${HOME}/github/powerline/fonts/install.sh
  fi
fi

echo -e "\e[32mCheck powerline fonts done.\e[0m"

# # Enviornment

echo -e "\e[34mChecking seed managed profile...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  cat <<EOF >${HOME}/.managed_profile
# # User
export ND_USER_HANDLE="${ND_USER_HANDLE}"
export ND_USER_DISPLAY_NAME="${ND_USER_DISPLAY_NAME}"
# # Editor
export EDITOR="vim"
EOF
  cat <<EOF >${HOME}/.managed_shrc
# # User
function me {
  printf "\$ND_USER_HANDLE"
}
EOF
  if ! grep -F -x 'source $HOME/.managed_profile' ${HOME}/.profile; then
    printf '\nsource $HOME/.managed_profile\n' >>${HOME}/.profile
  fi
  if ! grep -F -x 'source $HOME/.managed_profile' ${HOME}/.bash_profile; then
    printf '\nsource $HOME/.managed_profile\n' >>${HOME}/.bash_profile
  fi
  if ! grep -F -x 'source $HOME/.managed_profile' ${HOME}/.zprofile; then
    printf '\nsource $HOME/.managed_profile\n' >>${HOME}/.zprofile
  fi
  if ! grep -F -x 'source $HOME/.managed_shrc' ${HOME}/.bashrc; then
    printf '\nsource $HOME/.managed_shrc\n' >>${HOME}/.bashrc
  fi
  if ! grep -F -x 'source $HOME/.managed_shrc' ${HOME}/.zshrc; then
    printf '\nsource $HOME/.managed_shrc\n' >>${HOME}/.zshrc
  fi
fi

if [[ ${oslike} == "debian" ]]; then
  if ${wsl}; then
    cat <<EOF >>${HOME}/.managed_profile
# # WSL
export WINDOWS_HOST=$(ip route show | grep -i default | awk "{ print \$3 }")
alias open="wslview"
EOF
  fi
fi

echo -e "\e[32mCheck seed managed profile done.\e[0m"

# # Proxy



if [[ ${oslike} == "debian" ]]; then
  cat <<EOF >>${HOME}/.managed_profile
# # Proxy






EOF
fi

if [[ ${oslike} == "debian" ]]; then
  if grep -o -P -z "Host github.com[\n]?[\s]*ProxyCommand.*[\n]?" ${HOME}/.ssh/config; then
    echo -e "\e[33mFound ssh proxy to github.com for git, skip.\e[0m"
  else


  fi
fi



# # Tools

# ## System Packages

echo -e "\e[34mChecking system packages...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  sudo apt install -y clang
  sudo apt install -y clang-format
  sudo apt install -y default-jdk
  sudo apt install -y g++
  sudo apt install -y gcc
  sudo apt install -y git
  sudo apt install -y gitg
  sudo apt install -y p7zip-full
  sudo apt install -y p7zip-rar
  sudo apt install -y python3
  sudo apt install -y python3-pip
  sudo apt install -y rsync
fi

echo -e "\e[32mCheck system packages done.\e[0m"

# ## Snap

echo -e "\e[34mChecking snap tools...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  cat <<EOF >>${HOME}/.managed_profile
# # Snap
export PATH="/snap/bin:\$PATH"
EOF
fi

echo -e "\e[32mCheck snap tools done.\e[0m"

# ## Git

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

# ## Golang

echo -e "\e[34mChecking golang tools...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  sudo snap install --classic go
  cat <<EOF >>${HOME}/.managed_profile
# # Golang
export PATH="\$HOME/go/bin:\$PATH"
EOF
  go install github.com/bazelbuild/buildtools/buildifier@latest
fi

echo -e "\e[32mCheck golang tools done.\e[0m"

# ## Nvm

echo -e "\e[34mChecking nvm tools...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  set +eux
  bash -c "$(curl -fsSL https://raw.githubusercontent.com/nvm-sh/nvm/v0.40.0/install.sh)"
  # Manually load NVM
  export NVM_DIR="$HOME/.nvm"
  . "$NVM_DIR/nvm.sh"
  nvm install --lts
  nvm use --lts
  set -eux
  corepack enable
fi

echo -e "\e[32mCheck nvm tools done.\e[0m"

# ## Bazel

echo -e "\e[34mChecking bazelisk...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  npm list --global @bazel/bazelisk || npm install --global @bazel/bazelisk
fi

echo -e "\e[34mCheck bazelisk done.\e[0m"

# ## Node Tools

echo -e "\e[34mChecking node tools...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  npm list --global prettier || npm install --global prettier
  npm list --global typescript || npm install --global typescript
fi

echo -e "\e[32mCheck node tools done.\e[0m"

# Monorepo

echo -e "\e[34mChecking theseed monorepo...\e[0m"

if [[ ${oslike} == "debian" ]]; then
  mkdir -p ${HOME}/theseed

  if [[ -d ${HOME}/theseed/theseed.git && -d ${HOME}/theseed/main && -d ${HOME}/theseed/dev ]]; then
    echo -e "\e[33mFound existing theseed monorepo, skip clone.\e[0m"
  else
    if [[ -d ${HOME}/theseed/theseed.git || -d ${HOME}/theseed/main || -d ${HOME}/theseed/dev ]]; then
      echo "\e[31mFound old theseed monorepo, please backup and remove it:\e[33m
    rm -rf ${HOME}/theseed/dev
    rm -rf ${HOME}/theseed/main
    rm -rf ${HOME}/theseed/theseed.git
\e[0m"
      exit 1
    fi
    echo -e "\e[34mCloning theseed monorepo...\e[0m"
    git clone --bare --single-branch \
      --config "core.logallrefupdates=true" \
      --config "remote.origin.fetch=+refs/heads/*:refs/remotes/origin/*" \
      git@github.com:ndscm/theseed.git \
      ${HOME}/theseed/theseed.git
    git --git-dir ${HOME}/theseed/theseed.git worktree add -B main ${HOME}/theseed/main origin/main
    git --git-dir ${HOME}/theseed/theseed.git branch --track=direct base/dev origin/main
    git --git-dir ${HOME}/theseed/theseed.git branch --track=direct dev base/dev
    git --git-dir ${HOME}/theseed/theseed.git worktree add -B dev ${HOME}/theseed/dev
  fi
  git --git-dir ${HOME}/theseed/theseed.git config user.name "${ND_USER_DISPLAY_NAME}"
  git --git-dir ${HOME}/theseed/theseed.git config user.email "${ND_USER_HANDLE}@ndscm.com"

  cat <<EOF >>${HOME}/.managed_profile
# # Monorepo
export SEED_MONOREPO_HOME="\${HOME}/theseed"
export SEED_MONOREPO_GIT_DIR="\$SEED_MONOREPO_HOME/theseed.git"
export SEED_MAIN_HOME="\$SEED_MONOREPO_HOME/main"
export SEED_DEV_HOME="\$SEED_MONOREPO_HOME/dev"
EOF

  cd ${HOME}/theseed/main
  git fetch --all --prune
  git rebase origin/main
fi

echo -e "\e[34mCheck theseed monorepo done.\e[0m"

# # Shortcuts

if [[ ${oslike} == "debian" ]]; then
  cat <<EOF >>${HOME}/.managed_profile
# # Shortcuts
function main { cd \$SEED_MAIN_HOME; }
function dev { cd \$SEED_DEV_HOME; }
EOF
fi

if [[ ${oslike} == "debian" ]]; then
  cat <<EOF >>${HOME}/.managed_profile
# # Personal Profile
if [ -f ~/.personal_profile ]; then
  source \$HOME/.personal_profile
fi
EOF
fi

echo -e "\e[32mDone. Please restart the terminal to load new enviornments.\e[0m"
