#!/usr/bin/env bash

set -ex

echo "Installing Kustomize"
KUSTOMIZE=_bin/kustomize
KUSTOMIZE_VERSION=${KUSTOMIZE_VERSION:-3.8.1}

if [ -f ${KUSTOMIZE} ]; then
  rm ${KUSTOMIZE}
fi
mkdir -p _bin; cd _bin 
curl -s "https://raw.githubusercontent.com/\
kubernetes-sigs/kustomize/master/hack/install_kustomize.sh"  | bash -s ${KUSTOMIZE_VERSION}
cd -

# All the overlays to be built
DEPLOYMENT=${1:-sandbox test eks gcp azure}

KUSTOMIZE_OVERLAYS_ROOT=kustomize/overlays

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

for deployment in ${DEPLOYMENT}; do
    ${KUSTOMIZE} build ${KUSTOMIZE_OVERLAYS_ROOT}/${deployment} > ${DIR}/../deployment/${deployment}/flyte_generated.yaml
done

# This section is used by GitHub workflow to ensure that the generation step was run
if [ -n "$DELTA_CHECK" ]; then
  DIRTY=$(git status --porcelain)
  if [ -n "$DIRTY" ]; then
    echo "FAILED: kustomize code updated without commiting generated code."
    echo "Ensure make kustomize has run and all changes are committed."
    DIFF=$(git diff)
    echo "diff detected: $DIFF"
    DIFF=$(git diff --name-only)
    echo "files different: $DIFF"
    exit 1
  else
    echo "SUCCESS: Generated code is up to date."
  fi
fi
