#!/bin/sh

# Fix ownership of PostgreSQL data directory if it exists from a previous run
# (e.g., old bitnami PostgreSQL used uid 1001, embedded-postgres uses uid 999).
if [ -d /var/lib/flyte/storage/db ]; then
  chown -R 999:999 /var/lib/flyte/storage/db
fi

# Start embedded PostgreSQL in the background (must be running before k3s
# deploys the flyte-binary pod, which has a wait-for-db init container).
embedded-postgres &

# Wait for PostgreSQL to be ready before proceeding
while ! [ -f /tmp/embedded-postgres-ready ]; do
  sleep 0.5
done

flyte-devbox-bootstrap

# Wait for K3s to write kubeconfig to the staging path, rename the default
# context, add a flyte-devbox alias, then copy to the host-mounted output.
STAGING="${K3S_KUBECONFIG_OUTPUT:-/etc/rancher/k3s/k3s.yaml}"
(
  while ! [ -s "$STAGING" ]; do sleep 0.5; done
  # TODO: Remove flytev2-sandbox after all users have upgraded flyte-sdk.
  sed -i 's/: default/: flytev2-sandbox/g' "$STAGING"
  KUBECONFIG="$STAGING" kubectl config set-context flyte-devbox \
    --cluster=flytev2-sandbox --user=flytev2-sandbox 2>/dev/null || true
  if [ -n "${KUBECONFIG_FINAL:-}" ] && [ "$KUBECONFIG_FINAL" != "$STAGING" ]; then
    cp "$STAGING" "$KUBECONFIG_FINAL"
  fi
) &
