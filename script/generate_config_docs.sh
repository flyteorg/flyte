#!/usr/bin/env bash

set -ex

echo "Generating Flyte Configuration Documents"
CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
ROOT_DIR=${CUR_DIR}/..
OUTPUT_DIR="${ROOT_DIR}"/rsts/deployment/cluster_config/config.rst

GO111MODULE=on go get github.com/flyteorg/flyteadmin/cmd@config_docs

echo ".. _deployment-cluster-config-specification:" > "${ROOT_DIR}"/rsts/deployment/cluster_config/config.rst
echo "" >> "${ROOT_DIR}"/rsts/deployment/cluster_config/config.rst
${GOPATH:-~/go}/bin/flyteadmin config docs >> "${OUTPUT_DIR}"
#${GOPATH:-~/go}/bin/flyteapropeller config docs >> "${OUTPUT_DIR}"
#${GOPATH:-~/go}/bin/flytectl config docs >> "${OUTPUT_DIR}"
#${GOPATH:-~/go}/bin/flyteplugins config docs >> "${OUTPUT_DIR}"
