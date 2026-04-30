#!/usr/bin/env bash
# Boots the bundled flyte-devbox image as a single-container k3s cluster
# suitable for integration tests. Mirrors `make start` in
# docker/devbox-bundled/Makefile but headless and CI-friendly: writes
# kubeconfig to $PWD/.kube/kubeconfig, exports KUBECONFIG via $GITHUB_ENV,
# and waits for the flyte-binary pod to report Ready before returning.
set -euo pipefail

IMAGE="${DEVBOX_IMAGE:-flyte-devbox:ci}"
NAME="${DEVBOX_NAME:-flyte-devbox}"
KUBE_DIR="${KUBE_DIR:-$PWD/.kube}"
READY_TIMEOUT="${READY_TIMEOUT:-300}"

mkdir -p "$KUBE_DIR"
rm -f "$KUBE_DIR/kubeconfig"

docker run -d --rm --privileged --name "$NAME" \
  --add-host host.docker.internal:host-gateway \
  -e K3S_KUBECONFIG_OUTPUT=/.kube/kubeconfig \
  -v "$KUBE_DIR":/.kube \
  -p 6443:6443 \
  -p 30000:30000 \
  -p 30001:5432 \
  -p 30002:30002 \
  -p 30080:30080 \
  -p 30081:30081 \
  "$IMAGE"

echo "Waiting for kubeconfig (timeout ${READY_TIMEOUT}s)..."
deadline=$(( $(date +%s) + READY_TIMEOUT ))
until [ -s "$KUBE_DIR/kubeconfig" ]; do
  if [ "$(date +%s)" -gt "$deadline" ]; then
    echo "ERROR: kubeconfig not written within ${READY_TIMEOUT}s" >&2
    docker logs "$NAME" >&2 || true
    exit 1
  fi
  sleep 1
done
docker exec "$NAME" chown "$(id -u):$(id -g)" /.kube/kubeconfig

KUBECONFIG="$KUBE_DIR/kubeconfig"
export KUBECONFIG
if [ -n "${GITHUB_ENV:-}" ]; then
  echo "KUBECONFIG=$KUBECONFIG" >> "$GITHUB_ENV"
fi

echo "Waiting for flyte-binary pod to be Ready..."
kubectl wait --for=condition=Ready pod -n flyte \
  -l app.kubernetes.io/name=flyte-binary \
  --timeout="${READY_TIMEOUT}s"

echo "Devbox ready."
echo "  Connect API: http://localhost:30080"
echo "  rustfs S3:   http://localhost:30002"
echo "  Postgres:    localhost:30001"
