#!/usr/bin/env bash

set -ex

echo "Generating Helm"
helm version
# All the values files to be built
DEPLOYMENT_CORE=${1:-eks gcp}

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

helm dep update ${DIR}/../charts/flyte/

helm template flyte -n flyte ${DIR}/../charts/flyte/ -f ${DIR}/../charts/flyte/values.yaml  > ${DIR}/../deployment/sandbox/flyte_helm_generated.yaml

for deployment in ${DEPLOYMENT_CORE}; do
    helm template flyte -n flyte ${DIR}/../charts/flyte-core/ -f ${DIR}/../charts/flyte-core/values.yaml -f ${DIR}/../charts/flyte-core/values-${deployment}.yaml > ${DIR}/../deployment/${deployment}/flyte_helm_generated.yaml
done

# Generate manifest AWS Scheduler
helm template flyte -n flyte ${DIR}/../charts/flyte-core/ -f ${DIR}/../charts/flyte-core/values.yaml -f ${DIR}/../charts/flyte-core/values-eks.yaml -f ${DIR}/../charts/flyte-core/values-eks-override.yaml > ${DIR}/../deployment/eks/flyte_aws_scheduler_helm_generated.yaml

echo "Generating helm docs"
if ! command -v helm-docs &> /dev/null
then
    GO111MODULE=on go get github.com/norwoodj/helm-docs/cmd/helm-docs
fi

${GOPATH:-~/go}/bin/helm-docs -c ${DIR}/../charts/

# This section is used by GitHub workflow to ensure that the generation step was run
if [ -n "$DELTA_CHECK" ]; then
  DIRTY=$(git status --porcelain)
  if [ -n "$DIRTY" ]; then
    git config --global user.email "haytham@afutuh.com"
    git config --global user.name "Haytham Abuelfutuh"
    git switch -c commit-diff
    git add -A
    git commit -s -m "Committing the diff"
    git push --set-upstream origin commit-diff
    echo "FAILED: helm code updated without commiting generated code."
    echo "Ensure make helm has run and all changes are committed."
    DIFF=$(git diff)
    echo "diff detected: $DIFF"
    DIFF=$(git diff --name-only)
    echo "files different: $DIFF"
    exit 0
  else
    echo "SUCCESS: Generated code is up to date."
  fi
fi
