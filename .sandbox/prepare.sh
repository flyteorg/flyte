#!/usr/bin/env bash

set -e

PLATFORM="$(uname | tr '[:upper:]' '[:lower:]')"

get_kubectl() {
    local executable="$1/kubectl"
    local version="${KUBECTL_VERSION:-v1.20.2}"

    ( [ -f ${executable} ] && $(${executable} version 2> /dev/null | grep -q "GitVersion:\"${version}\"") ) || ( curl -Ls -o ${executable} https://dl.k8s.io/release/${version}/bin/${PLATFORM}/amd64/kubectl && chmod +x ${executable} )
}

# Get dependencies
mkdir -p .sandbox/bin
for dep in kubectl; do
    "get_${dep}" .sandbox/bin
done

# Build sandbox image
docker build \
    --build-arg DOCKER_VERSION="${DOCKER_VERSION:-20.10.3}" \
    --build-arg K3S_VERSION="${K3S_VERSION:-v1.20.2%2Bk3s1}" \
    -t "${FLYTE_SANDBOX_IMAGE:-flyte-sandbox:latest}" .sandbox
