#!/usr/bin/env bash

set -ex

DEPLOYMENT=${1:-sandbox test}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
KUSTOMIZE_IMAGE="lyft/kustomizer:v3.1.0"

for deployment in ${DEPLOYMENT}; do
    docker run -v "${DIR}/../kustomize":/kustomize "$KUSTOMIZE_IMAGE" kustomize build \
           "overlays/${deployment}/flyte" \
           > "${DIR}/../deployment/${deployment}/flyte_generated.yaml"
done
