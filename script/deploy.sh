set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

CONTEXT=$(kubectl config current-context)

if [ "$CLUSTER" == "docker-for-desktop" ] || [ "$CLUSTER" == "docker-desktop" ]; then
  DEPLOY_FILE="${DIR}/../deployment/sandbox/flyte_generated.yaml"
  kubectl apply -f "$DEPLOY_FILE"
else
  echo "$CLUSTER not supported - Sandbox deploys only allowed on 'docker-for-desktop' and 'docker-desktop' clusters"
  exit 1
fi
