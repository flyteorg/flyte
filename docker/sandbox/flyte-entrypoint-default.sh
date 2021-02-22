#!/bin/sh

set -e

trap 'pkill -P $$' EXIT

# Start k3s
echo "Starting k3s cluster..."
k3s server --no-deploy=traefik --no-deploy=servicelb --no-deploy=local-storage --no-deploy=metrics-server &> /var/log/k3s.log &
K3S_PID=$!
timeout 600 sh -c "until k3s kubectl explain deployment &> /dev/null; do sleep 1; done" || ( echo >&2 "Timed out while waiting for the Kubernetes cluster to start"; exit 1 )
echo "Done."

# Deploy flyte
echo "Deploying Flyte..."
k3s kubectl apply -f /flyteorg/share/flyte_generated.yaml
wait-for-flyte.sh
echo "Done."

wait ${K3S_PID}
