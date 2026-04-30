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

echo "Waiting for flyte namespace..."
until kubectl get ns flyte >/dev/null 2>&1; do
  if [ "$(date +%s)" -gt "$deadline" ]; then
    echo "ERROR: flyte namespace not created within ${READY_TIMEOUT}s" >&2
    kubectl get ns >&2 || true
    exit 1
  fi
  sleep 2
done

echo "Waiting for flyte-binary deployment to exist..."
until kubectl get deploy -n flyte flyte-binary >/dev/null 2>&1; do
  if [ "$(date +%s)" -gt "$deadline" ]; then
    echo "ERROR: flyte-binary deployment not created within ${READY_TIMEOUT}s" >&2
    kubectl get all -A >&2 || true
    exit 1
  fi
  sleep 2
done

remaining=$(( deadline - $(date +%s) ))
[ "$remaining" -lt 30 ] && remaining=30
echo "Waiting for flyte-binary rollout (timeout ${remaining}s)..."
kubectl rollout status deploy/flyte-binary -n flyte --timeout="${remaining}s"

# Bridge rustfs.flyte:9000 -> localhost:30002 (the rustfs NodePort).
# DataProxy mints signed URLs whose host is the in-cluster storage endpoint
# (http://rustfs.flyte:9000), which is unreachable from the runner. We add a
# /etc/hosts entry and a TCP forwarder so the SDK's PUT to the signed URL
# resolves to the published NodePort and lands on the rustfs pod.
if ! grep -q '[[:space:]]rustfs\.flyte\b' /etc/hosts; then
  echo "127.0.0.1 rustfs.flyte" | sudo tee -a /etc/hosts >/dev/null
fi
if ! command -v socat >/dev/null 2>&1; then
  sudo apt-get update -qq && sudo apt-get install -y -qq socat
fi
nohup socat TCP-LISTEN:9000,reuseaddr,fork TCP:127.0.0.1:30002 \
  >/tmp/rustfs-forward.log 2>&1 &
disown
forward_deadline=$(( $(date +%s) + 15 ))
until nc -z 127.0.0.1 9000 2>/dev/null; do
  if [ "$(date +%s)" -gt "$forward_deadline" ]; then
    echo "ERROR: rustfs.flyte:9000 forward did not open" >&2
    cat /tmp/rustfs-forward.log >&2 || true
    exit 1
  fi
  sleep 0.3
done

echo "Devbox ready."
echo "  Connect API: http://localhost:30080"
echo "  rustfs S3:   http://localhost:30002"
echo "  Postgres:    localhost:30001"
