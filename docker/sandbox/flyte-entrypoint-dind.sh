#!/bin/sh

set -euo pipefail

# Apply cgroup v2 hack
cgroup-v2-hack.sh

trap 'pkill -P $$' EXIT

monitor() {
    while : ; do
        for pid in $@ ; do
            kill -0 $pid &> /dev/null || exit 1
        done

        sleep 1
    done
}

# Start docker daemon
echo "Starting Docker daemon..."
dockerd &> /var/log/dockerd.log &
DOCKERD_PID=$!
timeout 600 sh -c "until docker info &> /dev/null; do sleep 1; done" || ( echo >&2 "Timed out while waiting for dockerd to start"; exit 1 )
echo "Done."

# Start k3s
echo "Starting k3s cluster..."
KUBERNETES_API_PORT=${KUBERNETES_API_PORT:-6443}
k3s server --docker --no-deploy=traefik --no-deploy=servicelb --no-deploy=local-storage --no-deploy=metrics-server --https-listen-port=${KUBERNETES_API_PORT} &> /var/log/k3s.log &
K3S_PID=$!
timeout 600 sh -c "until k3s kubectl explain deployment &> /dev/null; do sleep 1; done" || ( echo >&2 "Timed out while waiting for the Kubernetes cluster to start"; exit 1 )
echo "Done."

# Deploy flyte
echo "Deploying Flyte..."
helm install -n flyte -f /flyteorg/share/flyte/values-sandbox.yaml --create-namespace flyte /flyteorg/share/flyte --kubeconfig /etc/rancher/k3s/k3s.yaml

wait-for-flyte.sh

# Monitor running processes. Exit when the first process exits.
monitor ${DOCKERD_PID} ${K3S_PID}
