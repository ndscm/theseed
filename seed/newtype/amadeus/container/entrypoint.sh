#!/usr/bin/env bash
set -eux
set -o pipefail

# The container starts as root. In rootless podman this is the host service user
# mapped into the user namespace, which still holds CAP_CHOWN over its
# subordinate uid range; under docker it is real root. Either way root can fix
# the ownership of the mounted /home (non-recursively) without needing file
# capabilities. Once that is done we re-exec ourselves as the unprivileged
# amadeus user for the actual workload.
if [[ "$(id -u)" -eq 0 ]]; then
  mkdir -p /home/amadeus
  chown amadeus:amadeus /home/amadeus
  chmod 0750 /home/amadeus
  exec setpriv \
    --reuid amadeus \
    --regid amadeus \
    --init-groups \
    env "HOME=/home/amadeus" \
    "$0" "$@"
fi

if [[ -z "$(ls -A "/home/amadeus/")" ]]; then
  cp --recursive --no-target-directory /etc/skel/ "${HOME}/"
  cp /etc/zsh/newuser.zshrc.recommended /home/amadeus/.zshrc
  printf "%s\n" "export PATH=\"/home/amadeus/.local/bin:\$PATH\"" >>/home/amadeus/.zshrc
fi

# Run the command under an interactive zsh so that any CLI tools spawned from
# here (notably the Claude CLI and other developer tooling) see the same shell
# environment a real developer would have on their workstation: ~/.zshrc is
# sourced, PATH includes ~/.local/bin, aliases and completions are loaded, and
# tools that probe `$-` for an interactive shell behave as expected.
exec zsh -i -c 'exec "$@"' zsh "$@"
