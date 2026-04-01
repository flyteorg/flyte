#!/bin/sh

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
