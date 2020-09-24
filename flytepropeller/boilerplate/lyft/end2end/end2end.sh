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

# TODO: load all images
kind  load docker-image ${PROPELLER}

echo "Setup Kubectl"
kubectl cluster-info
kubectl get pods -n kube-system
echo "current-context:" $(kubectl config current-context)
echo "environment-kubeconfig:" ${KUBECONFIG}

pushd ${OUT}
# TODO: Only replace propeller if it's passed in
# TODO: Support replacing other images too
sed -i.bak -e "s_docker.io/lyft/flytepropeller:.*_${PROPELLER}_g" ${OUT}/kustomize/base/propeller/deployment.yaml
make kustomize
make end2end_execute
popd
