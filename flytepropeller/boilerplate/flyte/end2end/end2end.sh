#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'FLYTEORG/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/flyteorg/boilerplate/blob/master/Readme.rst

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

OUT="${DIR}/tmp"
rm -rf ${OUT}
git clone https://github.com/flyteorg/flyte.git "${OUT}"

pushd ${OUT}

if [ ! -z "$IMAGE" ]; 
then
  kind load docker-image ${IMAGE}
  if [ ${IMAGE_NAME} -eq "flytepropeller" ]
  then
    sed -i.bak -e "s_${IMAGE_NAME}:.*_${IMAGE}_g" ${OUT}/kustomize/base/propeller/deployment.yaml
  fi

  if [ ${IMAGE} -eq "flyteadmin" ]
  then
    sed -i.bak -e "s_${IMAGE_NAME}:.*_${IMAGE}_g" ${OUT}/kustomize/base/admindeployment/deployment.yaml
  fi
fi

make kustomize
# launch flyte end2end
kubectl apply -f "${OUT}/deployment/test/flyte_generated.yaml"
make end2end_execute
popd
