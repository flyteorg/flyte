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
file="/var/run/docker.pid"
if [ -f "$file" ] ; then
    rm "$file"
fi
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

FLYTE_VERSION=${FLYTE_VERSION:-latest}
if [[ $FLYTE_VERSION = "latest" ]]
then
  FLYTE_VERSION=$(curl --silent "https://api.github.com/repos/flyteorg/flyte/releases/latest" | jq -r .tag_name)
fi

# Deploy flyte
echo "Deploying Flyte..."
version=""
charts="/flyteorg/share/flyte"

if [[ $FLYTE_TEST = "release" ]]
then
  helm repo add flyteorg https://flyteorg.github.io/flyte
  helm fetch flyteorg/flyte --version=$FLYTE_VERSION
  version="--version $FLYTE_VERSION"
  charts="flyteorg/flyte"
fi

if [[ $FLYTE_TEST = "local" ]]
then
  helm dep update $charts
fi

helm upgrade -n flyte --create-namespace flyte $charts --kubeconfig /etc/rancher/k3s/k3s.yaml --install $version

wait-for-flyte.sh

# With flytectl sandbox --source flag, we mount the root volume to user source dir that will create helm & k8s cache specific directory.
# In Linux, These file belongs to root user that is different then current user
# In this case during fast serialization, Pyflyte will through error because of permission denied
rm -rf /root/.cache /root/.kube /root/.config

# Monitor running processes. Exit when the first process exits.
monitor ${DOCKERD_PID} ${K3S_PID}
