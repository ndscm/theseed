#!/bin/bash
set -eux
set -o pipefail

/opt/amadeus/gain-home

if [[ -z "$(ls -A "/home/amadeus/")" ]]; then
  cp --recursive --no-target-directory /etc/skel/ "${HOME}/"
  cp /etc/zsh/newuser.zshrc.recommended /home/amadeus/.zshrc
  printf "export PATH=\"/home/amadeus/.local/bin:\$PATH\"" >>/home/amadeus/.zshrc
  curl --fail --silent --show-error --location https://claude.ai/install.sh | bash
fi

# Run the command under an interactive zsh so that any CLI tools spawned from
# here (notably the Claude CLI and other developer tooling) see the same shell
# environment a real developer would have on their workstation: ~/.zshrc is
# sourced, PATH includes ~/.local/bin, aliases and completions are loaded, and
# tools that probe `$-` for an interactive shell behave as expected.
exec zsh -i -c 'exec "$@"' zsh "$@"
