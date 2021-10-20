#!/usr/bin/env bash

# Fail if any command fails and print what this script is doing
set -ex
# Also fail in the middle of a pipe if we ever have pipes
set -o pipefail
# Apply traps to all function calls
set -o errtrace

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

# Dump logs out because trying to upload to Travis artifacts is too hard.
# This is the naive solution, think of a better one in the future.
trap $DIR/print_logs.sh ERR

# This is a separate function so that we can potentially reuse in the future when we have more than one test
function wait_for_flyte_deploys() {
    SECONDS=0
    echo ""
    echo "waiting for flyte deploy to complete..."
    echo ""

    # wait for flyteadmin deployment to complete
    kubectl -n flyte rollout status deployment flyteadmin
    echo ""
    kubectl -n flyte rollout status deployment flytepropeller
    echo ""
    kubectl -n flyte rollout status deployment minio
    echo ""

    echo "Flyte deployed in $SECONDS seconds."
}

function timeout() { perl -e 'alarm shift; exec @ARGV' "$@"; }

function run_flyte_examples()
{
    echo $DIR
    # Launch test
    kubectl -n flyte create -f $DIR/tests/endtoend.yaml
    # Wait at most 20 minutes for things to pass
    timeout 1200 $DIR/test_monitor.sh
    return $?
}

wait_for_flyte_deploys

# In the future, to run additional tests against different containers, you can delete/reapply the K8s file
# to reset the cluster, wait for deployments again, and run additional functions.
run_flyte_examples
# Dump the logs from the run, automatically happens on error because of the trap, but we want this
# even for successful runs
# TODO: Move into the above function in the future
kubectl -n flyte logs endtoend

