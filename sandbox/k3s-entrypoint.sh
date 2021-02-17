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
dockerd &
DOCKERD_PID=$!
timeout 10 sh -c "until docker info &> /dev/null; do sleep 1; done"

# Start k3s
k3s server --docker --https-listen-port "${KUBERNETES_API_PORT:-6443}" --no-deploy=traefik --no-deploy=servicelb --no-deploy=local-storage --no-deploy=metrics-server &
K3S_PID=$!

# Monitor running processes. Exit when the first process exits.
monitor $DOCKERD_PID $K3S_PID
