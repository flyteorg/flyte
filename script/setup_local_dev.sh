#!/usr/bin/env bash

: "${ENSURE_K3D:="false"}" # will ensure k3d is installed if set to "true"
: "${ENSURE_KUBECTL:="false"}" # will ensure kubectl is installed if set to "true"
: "${ENSURE_HELM:="false"}" # will ensure helm is installed if set to "true"

: "${K3D_VERSION:="v5.4.1"}" # version of k3d to install, empty value uses latest one available
: "${K3D_INSTALL_URL:="https://raw.githubusercontent.com/rancher/k3d/main/install.sh"}" # URL to k3d installer script
: "${K3S_VERSION:="v1.21.1-k3s1"}" # version of k3s to run in k3d cluster, empty value uses default specified by k3d install
: "${K3D_CLUSTER_NAME:="flyte"}" # name of k3d cluster to be used
: "${K3D_KUBECONFIG_FILE_PATH:="${HOME}/.flyte/sandbox/kubeconfig"}" # file path to store kubeconfig file for k3d cluster at

: "${KUBECTL_VERSION:=""}" # version of kubectl to install, empty value uses latest available
: "${KUBECTL_INSTALL_URL:="https://dl.k8s.io/release/VERSION/bin/linux/amd64/kubectl"}" # URL to kubectl binary, include VERSION to be replaced with KUBECTL_VERSION
: "${KUBECTL_INSTALL_DIR:="/usr/local/bin"}" # local path to install kubectl to
: "${KUBECTL_VERIFY_SIGNATURE:="true"}" # verifies signature of kubectl binary downloaded if set to "true"

: "${HELM_VERSION:=""}" # version of helm to install, empty value uses latest v3 available
: "${HELM_INSTALL_URL:="https://raw.githubusercontent.com/helm/helm/main/scripts/get-helm-3"}" # URL to helm installer script
: "${HELM_K8S_NAMESPACE:="flyte"}" # k8s namespace to use with helm

: "${FLYTE_TIMEOUT:=600}"

set -e

HAS_CURL="$(type "curl" &> /dev/null && echo true || echo false)"
HAS_WGET="$(type "wget" &> /dev/null && echo true || echo false)"

if [ "${HAS_CURL}" != "true" ] && [ "${HAS_WGET}" != "true" ]; then
  echo "Either curl or wget is required"
  exit 1
fi

echo "Setting up local development environment"

if [ "${ENSURE_K3D}" == "true" ]; then
  echo -e "\nEnsuring k3d is installed"

  if [ "${HAS_CURL}" == "true" ]; then
    curl -sSL $K3D_INSTALL_URL | TAG=$K3D_VERSION bash
  else
    wget -q -O - $K3D_INSTALL_URL | TAG=$K3D_VERSION bash
  fi
else
  if type "k3d" &> /dev/null; then
    echo -e "\nUsing existing k3d install"
  else
    echo -e "\nk3d was not found in PATH, please install manually or run script with ENSURE_K3D=true"
    exit 1
  fi
fi

k3d version

if [ "${ENSURE_KUBECTL}" == "true" ]; then
  echo -e "\nEnsuring kubectl is installed"

  if [ "${HAS_CURL}" == "true" ]; then
    if [ -z "${KUBECTL_VERSION}" ]; then
      FULL_KUBECTL_INSTALL_URL="${KUBECTL_INSTALL_URL/VERSION/$(curl -sL https://dl.k8s.io/release/stable.txt)}"
    else
      FULL_KUBECTL_INSTALL_URL="${KUBECTL_INSTALL_URL/VERSION/${KUBECTL_VERSION}}"
    fi

    curl -sSLO $FULL_KUBECTL_INSTALL_URL

    if [ "${KUBECTL_VERIFY_SIGNATURE}" == "true" ]; then
      curl -sSLO "${FULL_KUBECTL_INSTALL_URL}.sha256"
    fi
  else
    if [ -z "${KUBECTL_VERSION}" ]; then
      FULL_KUBECTL_INSTALL_URL="${KUBECTL_INSTALL_URL/VERSION/$(wget -q -O - https://dl.k8s.io/release/stable.txt)}"
    else
      FULL_KUBECTL_INSTALL_URL="${KUBECTL_INSTALL_URL/VERSION/${KUBECTL_VERSION}}"
    fi

    wget -q $FULL_KUBECTL_INSTALL_URL

    if [ "${KUBECTL_VERIFY_SIGNATURE}" == "true" ]; then
      wget -q "${FULL_KUBECTL_INSTALL_URL}.sha256"
    fi
  fi

  if [ "${KUBECTL_VERIFY_SIGNATURE}" == "true" ]; then
    if ! echo "$(cat kubectl.sha256) kubectl" | sha256sum --check --status; then
      echo "Failed to verify signature of kubectl binary, aborting"
      rm kubectl kubectl.sha256 || true
      exit 1
    fi
  fi

  sudo install -o root -g root -m 0755 kubectl "${KUBECTL_INSTALL_DIR}/kubectl"
else
  if type "kubectl" &> /dev/null; then
    echo -e "\nUsing existing kubectl install"
  else
    echo -e "\nkubectl was not found in PATH, please install manually or run script with ENSURE_KUBECTL=true"
    exit 1
  fi
fi

kubectl version --client

if [ "${ENSURE_HELM}" == "true" ]; then
  echo -e "\nEnsuring helm is installed"

  if [ "${HAS_CURL}" == "true" ]; then
    curl -sSL $HELM_INSTALL_URL | DESIRED_VERSION=$HELM_VERSION bash
  else
    wget -q -O - $HELM_INSTALL_URL | DESIRED_VERSION=$HELM_VERSION bash
  fi
else
  if type "helm" &> /dev/null; then
    echo -e "\nUsing existing helm install"
  else
    echo -e "\nhelm was not found in PATH, please install manually or run script with ENSURE_HELM=true"
    exit 1
  fi
fi

helm version

if k3d cluster get $K3D_CLUSTER_NAME &> /dev/null; then
  echo -e "\nUsing existing k3d cluster ${K3D_CLUSTER_NAME}"

  k3d cluster start $K3D_CLUSTER_NAME
else
  echo -e "\nCreating k3d cluster ${K3D_CLUSTER_NAME}"

  IMAGE_ARG=""
  if [ -n "${K3S_VERSION}" ]; then
    IMAGE_ARG="--image rancher/k3s:${K3S_VERSION}"
  fi

  # shellcheck disable=SC2086
  k3d cluster create -p "30082:30082@server:0" -p "30084:30084@server:0" -p "30088:30088@loadbalancer" -p "30089:30089@server:0" ${IMAGE_ARG} $K3D_CLUSTER_NAME
fi

if [ -f "${K3D_KUBECONFIG_FILE_PATH}" ]; then
  echo -e "\nUsing existing k3d kubeconfig at ${K3D_KUBECONFIG_FILE_PATH}"
else
  echo -e "\nExporting k3d kubeconfig to ${K3D_KUBECONFIG_FILE_PATH}"
  k3d kubeconfig get $K3D_CLUSTER_NAME > "${K3D_KUBECONFIG_FILE_PATH}"
  chmod 600 "${K3D_KUBECONFIG_FILE_PATH}"
fi

echo -e "\nSetting kubeconfig and kubectl context"

export KUBECONFIG=$KUBECONFIG:"${K3D_KUBECONFIG_FILE_PATH}"
kubectl config set-context $K3D_CLUSTER_NAME

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
HELM_CHART="${DIR}/../charts/flyte-deps"

echo -e "\nUpdating helm dependencies"
helm dep update "${HELM_CHART}"

echo -e "\nApplying helm chart for flyte dependencies"
helm upgrade -n flyte --create-namespace $HELM_K8S_NAMESPACE "${HELM_CHART}" --kubeconfig "${K3D_KUBECONFIG_FILE_PATH}" --install --wait

echo -e "\nWaiting for helm chart rollouts to start"
timeout "$FLYTE_TIMEOUT" sh -c "until kubectl get namespace ${HELM_K8S_NAMESPACE} &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the ${HELM_K8S_NAMESPACE} namespace to be created"; exit 1 )
timeout "$FLYTE_TIMEOUT" sh -c "until kubectl rollout status deployment minio -n ${HELM_K8S_NAMESPACE}  &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the minio rollout to be created"; exit 1 )
timeout "$FLYTE_TIMEOUT" sh -c "until kubectl rollout status deployment postgres -n ${HELM_K8S_NAMESPACE}  &> /dev/null; do sleep 1; done"  || ( echo >&2 "Timed out while waiting for the postgres rollout to be created"; exit 1 )

echo -e "\nWaiting for Flyte dependencies deployment to start"
kubectl wait --for=condition=available deployment/minio deployment/postgres -n $HELM_K8S_NAMESPACE --timeout=5m &> /dev/null || ( echo >&2 "Timed out while waiting for the Flyte deployment to start"; exit 1 )

# Create directory to store certificate
mkdir -p /tmp/k8s-webhook-server/serving-certs

echo -e "\n*** Successfully set up local development environment."
echo -e "*** Run \"flyte start --config flyte-single-binary-local.yaml\" to start your local flyte installation.\n"
