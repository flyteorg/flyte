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

FLYTE_VERSION=${FLYTE_VERSION:-latest}
if [[ $FLYTE_VERSION = "latest" ]]
then
  FLYTE_VERSION=$(curl --silent "https://api.github.com/repos/flyteorg/flyte/releases/latest" | jq -r .tag_name)
fi

# Deploy flyte
echo "Deploying Flyte..."
helm repo add flyteorg https://flyteorg.github.io/flyte
helm repo update
helm fetch flyteorg/flyte --version=$FLYTE_VERSION
helm upgrade -n flyte --create-namespace flyte flyteorg/flyte --kubeconfig /etc/rancher/k3s/k3s.yaml --install --version $FLYTE_VERSION

wait-for-flyte.sh

wait ${K3S_PID}
