#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

. seed/devprod/smd/workstation/maintain-01-os.sh
. seed/devprod/smd/workstation/maintain-02-identity.sh
. seed/devprod/smd/workstation/maintain-11-basic-packages.sh
. seed/devprod/smd/workstation/maintain-12-ssh-identity.sh
. seed/devprod/smd/workstation/maintain-21-shell.sh
. seed/devprod/smd/workstation/maintain-22-font.sh
. seed/devprod/smd/workstation/maintain-31-managed-profile.sh
. seed/devprod/smd/workstation/maintain-32-managed-proxy.sh
. seed/devprod/smd/workstation/maintain-41-system-packages.sh
. seed/devprod/smd/workstation/maintain-42-snap.sh
. seed/devprod/smd/workstation/maintain-43-scm.sh
. seed/devprod/smd/workstation/maintain-44-go.sh
. seed/devprod/smd/workstation/maintain-45-node.sh
. seed/devprod/smd/workstation/maintain-46-bazel.sh
. seed/devprod/smd/workstation/maintain-47-python.sh
. seed/devprod/smd/workstation/maintain-48-container.sh
. seed/devprod/smd/workstation/maintain-51-monorepo.sh
. seed/devprod/smd/workstation/maintain-61-shortcut.sh
. seed/devprod/smd/workstation/maintain-99-finalize.sh
