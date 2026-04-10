#!/bin/sh

set -o errexit
set -o nounset

LOGFILE="/var/log/k3d-entrypoints_$(date "+%y%m%d%H%M%S").log"

touch "$LOGFILE"

echo "[$(date -Iseconds)] Running k3d entrypoints..." >> "$LOGFILE"

# Redirect K3S_KUBECONFIG_OUTPUT to a staging path so K3s doesn't write
# directly to the host-mounted path. The bootstrap entrypoint script will
# post-process the staging file (rename context, add aliases) then copy
# to the real output path — so the host only sees the finished version.
if [ -n "${K3S_KUBECONFIG_OUTPUT:-}" ]; then
  KUBECONFIG_FINAL="$K3S_KUBECONFIG_OUTPUT"
  export KUBECONFIG_FINAL
  export K3S_KUBECONFIG_OUTPUT="/tmp/k3s-kubeconfig-staging"
fi

for entrypoint in /bin/k3d-entrypoint-*.sh ; do
  echo "[$(date -Iseconds)] Running $entrypoint"  >> "$LOGFILE"
  "$entrypoint"  >> "$LOGFILE" 2>&1 || exit 1
done

echo "[$(date -Iseconds)] Finished k3d entrypoint scripts!" >> "$LOGFILE"

exec /bin/k3s "$@"
