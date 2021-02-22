#!/bin/sh

set -e

# Ensure cluster is up and running
timeout 180 sh -c "until k3s kubectl explain deployment &> /dev/null; do sleep 1; done"

# Wait for flyte deployment
k3s kubectl wait --for=condition=available deployment/datacatalog deployment/flyteadmin deployment/flyteconsole deployment/flytepropeller -n flyte --timeout=10m

timeout 180 sh -c 'until [[ $(k3s kubectl get daemonset envoy -n flyte -o jsonpath="{.status.numberReady}") -eq 1 ]]; do sleep 1; done'
