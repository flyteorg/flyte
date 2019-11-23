set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

CONTEXT=$(kubectl config current-context)
DEPLOY_FILE="${DIR}/../deployment/sandbox/flyte_generated.yaml"

function run_kubectl {
  kubectl apply -f "$DEPLOY_FILE"
}

if [ "$CONTEXT" == "docker-for-desktop" ] || [ "$CLUSTER" == "docker-desktop" ]; then
  run_kubectl
else
  echo "$CONTEXT is not a docker-desktop cluster"
  read -p "  Are you sure you want to continue? " answer
    if [ "$answer" != "${answer#[Yy]}" ] ;then
        run_kubectl
        exit 0
    else
        echo Exiting without doing anything
        exit 1
    fi
fi
