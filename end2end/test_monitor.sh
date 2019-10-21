#!/usr/bin/env bash

set -ex
set -o pipefail

function pod_succeeded()
{
    kubectl -n flyte get pod endtoend -o=custom-columns=STATUS:.status.phase | grep Succeeded > /dev/null
    return $?
}

function get_status()
{
    kubectl -n flyte get pod endtoend -o=custom-columns=STATUS:.status.phase | grep -v STATUS
}

function wait_until_terminal_state()
{
    status=$(get_status)
    while [[ "$status" != "Succeeded" && "$status" != "Failed" ]]; do
        echo "Status is still $status"
        sleep 5
        status=$(get_status)
    done
}

function watch_pod()
{
    wait_until_terminal_state
    return $(pod_succeeded)
}

watch_pod

