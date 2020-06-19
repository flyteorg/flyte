#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

OUT="${DIR}/tmp"
rm -rf ${OUT}
git clone https://github.com/lyft/flyte.git "${OUT}"

echo "Loading github docker images into 'kind' cluster to workaround this issue: https://github.com/containerd/containerd/issues/3291#issuecomment-631746985"
docker login --username ${DOCKER_USERNAME} --password ${DOCKER_PASSWORD} docker.pkg.github.com

pushd ${OUT}

if [ ! -z "$PROPELLER" ]; then
  docker pull docker.pkg.github.com/${PROPELLER}
  kind load docker-image docker.pkg.github.com/${PROPELLER}
  sed -i.bak -e "s_docker.io/lyft/flytepropeller:.*_docker.pkg.github.com/${PROPELLER}_g" ${OUT}/kustomize/base/propeller/deployment.yaml
fi

if [ ! -z "$ADMIN" ]; then
  docker pull docker.pkg.github.com/${ADMIN}
  kind load docker-image docker.pkg.github.com/${ADMIN}
  sed -i.bak -e "s_docker.io/lyft/flyteadmin:.*_docker.pkg.github.com/${ADMIN}_g" ${OUT}/kustomize/base/propeller/deployment.yaml
fi

make kustomize
make end2end_execute
popd
