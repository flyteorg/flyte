#!/bin/sh

set -e

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
timeout 180 sh -c "until docker info &> /dev/null; do sleep 1; done"
echo "Done."

# Start k3s
echo "Starting k3s cluster..."
k3s server --docker --no-deploy=traefik --no-deploy=servicelb --no-deploy=local-storage --no-deploy=metrics-server &> /var/log/k3s.log &
K3S_PID=$!
timeout 180 sh -c "until k3s kubectl explain deployment &> /dev/null; do sleep 1; done"
echo "Done."

# Deploy flyte
echo "Deploying Flyte..."
k3s kubectl apply -f /flyteorg/share/flyte_generated.yaml
wait-for-flyte.sh
echo "Done."

# Monitor running processes. Exit when the first process exits.
monitor ${DOCKERD_PID} ${K3S_PID}
