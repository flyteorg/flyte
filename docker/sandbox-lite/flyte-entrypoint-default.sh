#!/bin/sh

set -euo pipefail

# Apply cgroup v2 hack
cgroup-v2-hack.sh

trap 'pkill -P $$' EXIT

# Start k3s
echo "Starting k3s cluster..."
KUBERNETES_API_PORT=${KUBERNETES_API_PORT:-6443}
k3s server --no-deploy=traefik --no-deploy=servicelb --no-deploy=local-storage --no-deploy=metrics-server --https-listen-port=${KUBERNETES_API_PORT} &> /var/log/k3s.log &
K3S_PID=$!
timeout 600 sh -c "until k3s kubectl explain deployment &> /dev/null; do sleep 1; done" || ( echo >&2 "Timed out while waiting for the Kubernetes cluster to start"; exit 1 )
echo "Done."

# Deploy flyte
echo "Deploying Flyte..."
charts="/flyteorg/share/flyte"
helm dep update $charts
helm install -n flyte --create-namespace flyte $charts --kubeconfig /etc/rancher/k3s/k3s.yaml
k3s kubectl create namespace flytesnacks-development
k3s kubectl wait --for=condition=available deployment/minio deployment/postgres -n flyte --timeout=5m || ( echo >&2 "Timed out while waiting for the Flyte deployment to start"; exit 1 )
k3s kubectl port-forward svc/postgres -n flyte 5432:5432 &>/dev/null &
# Wait for Postgres port-forwarding to work because it doesn't start immediately
sleep 3
flyte start --config /flyteorg/share/flyte.yaml &
FLYTE_PID=$!
wait -n ${FLYTE_PID} ${K3S_PID}
