#!/usr/bin/env bash

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
FLYTEPROPELLER_IMAGE="lyft/flytepropeller:522bc92d4a45aab7163f2f5d5b0bf1594edfae11"

docker run -v "${DIR}/../kustomize":/kustomize -v "${DIR}/../crd_validation":/crd_validation "$FLYTEPROPELLER_IMAGE" bin/build-tool \
    crd-validation --base-crd /kustomize/base/wf_crd/wf_crd.yaml > ${DIR}/../kustomize/base/wf_crd/wf_crd_with_validation_generated.yaml

# TODO: add a command in flytepropeller's build-tool binary for checking the equivalency
