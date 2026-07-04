#!/usr/bin/env bash
set -eux
set -o pipefail

playpen_user="${1:-"christina"}"
init_process="${2:-""}"

if ! id "${playpen_user}" &>/dev/null; then
  useradd --shell /usr/bin/zsh --create-home "${playpen_user}"
fi

if [[ ! -f "/home/${playpen_user}/.zshrc" ]]; then
  cp /etc/zsh/newuser.zshrc.recommended "/home/${playpen_user}/.zshrc"
  chown "${playpen_user}:${playpen_user}" "/home/${playpen_user}/.zshrc"
  printf '\n# Path\nexport PATH="/home/%s/.local/bin:${PATH}"\n' "${playpen_user}" \
    >>"/home/${playpen_user}/.zshrc"
fi

if [[ "${init_process}" == "systemd" ]]; then
  exec /lib/systemd/systemd
else
  exec tail -f /dev/null
fi
