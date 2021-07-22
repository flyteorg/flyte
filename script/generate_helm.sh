#!/usr/bin/env bash

set -ex

echo "Generating Helm"
helm version
# All the values files to be built
DEPLOYMENT=${1:-sandbox eks gcp}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

helm dep update ${DIR}/../charts/flyte/

for deployment in ${DEPLOYMENT}; do
    helm template flyte -n flyte ${DIR}/../charts/flyte/ -f ${DIR}/../charts/flyte/values-${deployment}.yaml > ${DIR}/../deployment/${deployment}/flyte_helm_generated.yaml
done

echo "Generating helm docs"
if ! command -v helm-docs &> /dev/null
then
    GO111MODULE=on go get github.com/norwoodj/helm-docs/cmd/helm-docs
fi

${GOPATH:-~/go}/bin/helm-docs -t ${DIR}/../charts/flyte/README.md.gotmpl ${DIR}/../charts/flyte/
${GOPATH:-~/go}/bin/helm-docs -t ${DIR}/../charts/flyte-core/README.md.gotmpl ${DIR}/../charts/flyte-core/

# This section is used by GitHub workflow to ensure that the generation step was run
if [ -n "$DELTA_CHECK" ]; then
  DIRTY=$(git status --porcelain)
  if [ -n "$DIRTY" ]; then
    echo "FAILED: helm code updated without commiting generated code."
    echo "Ensure make helm has run and all changes are committed."
    DIFF=$(git diff)
    echo "diff detected: $DIFF"
    DIFF=$(git diff --name-only)
    echo "files different: $DIFF"
    exit 1
  else
    echo "SUCCESS: Generated code is up to date."
  fi
fi
