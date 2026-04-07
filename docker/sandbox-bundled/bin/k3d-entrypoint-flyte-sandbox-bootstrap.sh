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

flyte-sandbox-bootstrap

KUBECONFIG_PATH="${K3S_KUBECONFIG_OUTPUT:-/etc/rancher/k3s/k3s.yaml}"
(
  while ! [ -s "$KUBECONFIG_PATH" ]; do sleep 1; done
  sed -i 's/: default/: flytev2-sandbox/g' "$KUBECONFIG_PATH"
) &
