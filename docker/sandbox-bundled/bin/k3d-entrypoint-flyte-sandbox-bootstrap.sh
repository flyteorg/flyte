#!/bin/sh

# Fix ownership of PostgreSQL data directory if it exists from a previous run
# (e.g., old bitnami PostgreSQL used uid 1001, embedded-postgres uses uid 999).
if [ -d /var/lib/flyte/storage/db ]; then
  chown -R 999:999 /var/lib/flyte/storage/db
fi

# Start embedded PostgreSQL in the background. The flyte-binary pod has a
# wait-for-db init container, so postgres doesn't need to be ready before
# K3s starts — it just needs to be running by the time the pod is scheduled.
embedded-postgres &

# Render and write the K3s auto-deploy manifest. This only does template
# variable substitution (HOST_GATEWAY_IP, NODE_IP) — no postgres needed.
# Runs synchronously so the manifest is in place before K3s starts.
flyte-sandbox-bootstrap

KUBECONFIG_PATH="${K3S_KUBECONFIG_OUTPUT:-/etc/rancher/k3s/k3s.yaml}"
(
  while ! [ -s "$KUBECONFIG_PATH" ]; do sleep 1; done
  sed -i 's/: default/: flytev2-sandbox/g' "$KUBECONFIG_PATH"
) &
