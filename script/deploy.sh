set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

DEPLOY_FILE="${DIR}/../deployment/sandbox/flyte_generated.yaml"
kubectl apply -f "$DEPLOY_FILE"
