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

  # Give the unprivileged amadeus user a writable, controller-delegated cgroup-v2
  # subtree so the nested rootless podman can create cgroups for --systemd=always
  # playpen containers. This is what `systemd --user` does for a login session on
  # the host; we do it by hand because this container runs amadeus-server as PID
  # 1, not systemd. Requires a writable cgroup mount (--privileged
  # --systemd=always in the quadlet).
  #
  # cgroup v2's delegation-containment rule only allows migrating a process when
  # the writer can write the cgroup.procs of the *common ancestor* of the source
  # and destination cgroups. So amadeus-server (and the podman it spawns) must run
  # *inside* the delegated subtree, with container cgroups created as siblings of
  # it under a shared amadeus-owned parent -- not in a sibling tree. We therefore
  # delegate /amadeus to the amadeus user, run every process under the
  # /amadeus/init leaf, and point the nested podman at --cgroup-parent=/amadeus.
  # (cgroup v2 also forbids a cgroup from holding processes while it delegates
  # controllers, which is why processes live in the init leaf and never in
  # /amadeus itself.)
  if [[ -w /sys/fs/cgroup/cgroup.procs ]]; then
    mkdir -p /sys/fs/cgroup/amadeus/init
    for pid in $(cat /sys/fs/cgroup/cgroup.procs); do
      echo "${pid}" >/sys/fs/cgroup/amadeus/init/cgroup.procs 2>/dev/null || true
    done
    for controller in $(cat /sys/fs/cgroup/cgroup.controllers); do
      echo "+${controller}" >/sys/fs/cgroup/cgroup.subtree_control 2>/dev/null || true
    done
    for controller in $(cat /sys/fs/cgroup/amadeus/cgroup.controllers); do
      echo "+${controller}" >/sys/fs/cgroup/amadeus/cgroup.subtree_control 2>/dev/null || true
    done
    chown -R amadeus:amadeus /sys/fs/cgroup/amadeus
  fi

  if [[ -d "/playpen/home" ]]; then
    chown amadeus:amadeus /playpen/home
  fi

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

if [[ -w "/sys/fs/cgroup/amadeus/cgroup.procs" ]]; then
  podman load --input /opt/amadeus/playpen.tar
fi

exec "$@"
