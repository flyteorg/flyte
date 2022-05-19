#!/bin/sh

set -e

echo "Waiting for Flyte to become ready..."

FLYTE_TIMEOUT=${FLYTE_TIMEOUT:-600}

# Ensure cluster is up and running. We don't need a timeout here, since the container
# itself will exit with the appropriate error message if the kubernetes cluster is not
# up within the specified timeout.
until k3s kubectl explain deployment &> /dev/null; do sleep 1; done

# Wait for Flyte namespace to be created. This is necessary for the next step.
timeout $FLYTE_TIMEOUT sh -c "until k3s kubectl get namespace flyte &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the Flyte namespace to be created"; exit 1 )

# Wait for Flyte deployment to be created. This is necessary for the next step.
timeout $FLYTE_TIMEOUT sh -c "until k3s kubectl rollout status deployment datacatalog -n flyte  &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the datacatalog rollout to be created"; exit 1 )
timeout $FLYTE_TIMEOUT sh -c "until k3s kubectl rollout status deployment flyteadmin -n flyte  &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the flyteadmin rollout to be created"; exit 1 )
timeout $FLYTE_TIMEOUT sh -c "until k3s kubectl rollout status deployment flyteconsole -n flyte  &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the flyteconsole rollout to be created"; exit 1 )
timeout $FLYTE_TIMEOUT sh -c "until k3s kubectl rollout status deployment flytepropeller -n flyte  &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the flytepropeller rollout to be created"; exit 1 )
timeout $FLYTE_TIMEOUT sh -c "until k3s kubectl rollout status deployment flyte-deps-contour-contour -n flyte  &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the flytepropeller rollout to be created"; exit 1 )

# Wait for flyte deployment
k3s kubectl wait --for=condition=available deployment/datacatalog deployment/flyteadmin deployment/flyteconsole deployment/flytepropeller deployment/flyte-deps-contour-contour -n flyte --timeout=10m || ( echo >&2 "Timed out while waiting for the Flyte deployment to start"; exit 1 )

echo "Flyte is ready! Flyte UI is available at http://localhost:30081/console."
