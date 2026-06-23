#!/usr/bin/env bash
set -eux
set -o pipefail
cd "$(dirname "${BASH_SOURCE[0]}")/../../../.."

maintain_scopes="${1:-""}"
export maintain_scopes

. seed/devprod/smd/workstation/maintain-00-os.sh
. seed/devprod/smd/workstation/maintain-01-system-license.sh
. seed/devprod/smd/workstation/maintain-02-user-identity.sh
. seed/devprod/smd/workstation/maintain-10-system-basic-packages.sh
. seed/devprod/smd/workstation/maintain-11-user-basic-packages.sh
. seed/devprod/smd/workstation/maintain-12-user-ssh-identity.sh
. seed/devprod/smd/workstation/maintain-20-system-shell.sh
. seed/devprod/smd/workstation/maintain-21-user-shell.sh
. seed/devprod/smd/workstation/maintain-22-user-font.sh
. seed/devprod/smd/workstation/maintain-31-user-managed-profile.sh
. seed/devprod/smd/workstation/maintain-32-user-managed-proxy.sh
. seed/devprod/smd/workstation/maintain-40-system-packages.sh
. seed/devprod/smd/workstation/maintain-42-user-snap.sh
. seed/devprod/smd/workstation/maintain-43-user-scm.sh
. seed/devprod/smd/workstation/maintain-44-user-go.sh
. seed/devprod/smd/workstation/maintain-45-user-node.sh
. seed/devprod/smd/workstation/maintain-46-user-bazel.sh
. seed/devprod/smd/workstation/maintain-47-user-python.sh
. seed/devprod/smd/workstation/maintain-48-user-container.sh
. seed/devprod/smd/workstation/maintain-51-user-monorepo.sh
. seed/devprod/smd/workstation/maintain-61-user-shortcut.sh
. seed/devprod/smd/workstation/maintain-99-user-finalize.sh
