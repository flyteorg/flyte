#!/usr/bin/env bash

set -ex

echo "Install Kustomize"
curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash

DEPLOYMENT=${1:-sandbox test eks gcp}
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

for deployment in ${DEPLOYMENT}; do
    kustomize build kustomize/overlays/${deployment} > ${DIR}/../deployment/${deployment}/flyte_generated.yaml
done
