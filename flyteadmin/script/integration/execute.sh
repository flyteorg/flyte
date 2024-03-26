#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# TODO: load all images
docker tag ${IMAGE} "flyteadmin:test"
kind load docker-image flyteadmin:test

# start flyteadmin and dependencies
kubectl apply -f "${DIR}/k8s/integration.yaml"

# This is a separate function so that we can potentially reuse in the future when we have more than one test
function wait_for_flyte_deploys() {
    SECONDS=0
    echo ""
    echo "waiting for flyte deploy to complete..."
    echo ""

    # wait for flyteadmin deployment to complete
    kubectl -n flyte rollout status deployment flyteadmin
    echo ""

    echo "Flyte deployed in $SECONDS seconds."
}

wait_for_flyte_deploys

## get the name of the flyteadmin pod
POD_NAME=$(kubectl get pods -n flyte --field-selector=status.phase=Running -o go-template="{{range .items}}{{.metadata.name}}:{{end}}" | tr ":" "\n" | grep flyteadmin)

echo $POD_NAME

# launch the integration tests
kubectl exec -it -n flyte "$POD_NAME" -- make -C /go/src/github.com/flyteorg/flyteadmin integration
