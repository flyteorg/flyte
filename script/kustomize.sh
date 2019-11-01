set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
KUSTOMIZE_IMAGE="lyft/kustomizer:v3.1.0"

# Flyte test
docker run -v "${DIR}/../kustomize":/kustomize "$KUSTOMIZE_IMAGE" kustomize build overlays/test/flyte > "${DIR}/../deployment/test/flyte_generated.yaml"

# Flyte Sandbox
docker run -v "${DIR}/../kustomize":/kustomize "$KUSTOMIZE_IMAGE" kustomize build overlays/sandbox/flyte > "${DIR}/../deployment/sandbox/flyte_generated.yaml"

# canonical GCP deployment on GKE
docker run -v "${DIR}/../kustomize":/kustomize "$KUSTOMIZE_IMAGE" kustomize build overlays/gke-single-cluster/flyte > "${DIR}/../deployment/gke-single-cluster/flyte_generated.yaml"
