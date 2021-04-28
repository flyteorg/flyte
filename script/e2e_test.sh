#!/bin/bash

set -e

RUNNER=$1
PRIOROTY=$2
FLYTE_VERSION=$3
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
BASEDIR="${DIR}/.."

FLYTECTL_HOST="127.0.0.1:30086"


function register_workflow {
  FLYTESNACK_VERSION=$(curl --silent "https://api.github.com/repos/flyteorg/flytesnacks/releases/latest" | jq -r .tag_name)
  FLYTECTL_PROJECT_FLAGS="-d development -p flytesnacks ----admin.endpoint=${FLYTECTL_HOST}"
  # Download the flytesnacks test manifest
  curl https://github.com/flyteorg/flytesnacks/releases/download/${FLYTESNACK_VERSION}/flyte_tests_manifest.json -o ${BASEDIR}/flyte_tests_manifest.json
  for row in $(cat ${BASEDIR}/flyte_tests_manifest.json | jq -c '.[]'); do
    if [[ $(echo ${row} | jq -r '.priority') -eq $1 || $1 -eq "all" ]]
    then
        flytectl register files  https://github.com/evalsocket/flytesnacks/releases/download/$(FLYTESNACK_VERSION)/flytesnacks-$(echo ${row} | jq -r '.name').tgz ${FLYTECTL_PROJECT_FLAGS} --archive
    fi
  done
}

function create_sandbox {
  git clone git@github.com:flyteorg/flytesnacks.git ${BASEDIR}/flytesnacks
  cd ${BASEDIR}/flytesnacks
  make start
  make kubectl-config

  FLYTECTL_HOST="127.0.0.1:30086"
  FLYTE_VERSION=$(curl --silent "https://api.github.com/repos/flyteorg/flyte/releases/latest" | jq -r .tag_name)
  kubectl apply -f https://github.com/flyteorg/flyte/releases/download/${FLYTE_VERSION}/flyte_sandbox_manifest.yaml
}

function create_cluster {
  echo "Work needs to be done here"
  # TODO: Need to setup management cluster with CAPA & CAPG provider
  # TODO: Setup management cluster kubeconfig config and deploy cluster using cluster-api
  if [[ $1 == "capg" ]]; then
    echo "Deploy GCP Cluster CR"
  fi
  if [[ $1 == "capa" ]]; then
    echo "Deploy AWS Cluster CR"
  fi

  FLYTE_VERSION=$(curl --silent "https://api.github.com/repos/flyteorg/flyte/releases/latest" | jq -r .tag_name)
  kubectl apply -f https://github.com/flyteorg/flyte/releases/download/${FLYTE_VERSION}/flyte_sandbox_manifest.yaml
}

function tear_downs {
  cd ${BASEDIR}/flytesnacks
  make teardown
}

function run_e2e {
  FLYTESNACK_VERSION=$(curl --silent "https://api.github.com/repos/flyteorg/flytesnacks/releases/latest" | jq -r .tag_name)

  # Download the flytesnacks test manifest
  curl https://github.com/flyteorg/flytesnacks/releases/download/${FLYTESNACK_VERSION}/flyte_tests_manifest.json -o ${BASEDIR}/flyte_tests_manifest.json
  for row in $(cat ${BASEDIR}/flyte_tests_manifest.json | jq -c '.[]'); do
    if [[ $(echo ${row} | jq -r '.priority') -eq $1 || $1 -eq "all" ]]; then
        # TODO: Run test for specific priority
        echo "Run test for priority #{$1}"
    fi
  done
}

contains() {
    string="$1"
    substring="$2"

    if echo "$string" | $(type -p ggrep grep | head -1) -F -- "$substring" >/dev/null; then
        return 0    # $substring is in $string
    else
        return 1    # $substring is not in $string
    fi
}


function e2e_test {
  if [[ $RUNNER == "ubuntu-latest" ]]; then
    create_sandbox
  else
    if [[ $RUNNER == *"aws"* ]]; then
        create_cluster "capa"
    else
        create_cluster "capg"
    fi
  fi
  register_workflow $PRIOROTY
  run_e2e $PRIOROTY

  if [[ $RUNNER == "ubuntu-latest" ]]; then
    tear_downs_sandbox
  else
    tear_downs_cluster
  fi

}

e2e_test $RUNNER $PRIOROTY
