#!/usr/bin/env bash

set -e

echo "Generating Flyte Configuration Documents"
CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
ROOT_DIR=${CUR_DIR}/..
OUTPUT_DIR="${ROOT_DIR}"/rsts/deployment/cluster_config/config.rst

GO111MODULE=on go get github.com/flyteorg/flyteadmin/cmd@config_docs

output_config () {
echo "
$1
===============================
" >> "${OUTPUT_DIR}"
"${GOPATH:-~/go}"/bin/$2 config docs >> "${OUTPUT_DIR}"
}

echo ".. _deployment-cluster-config-specification:

###################################
Flyte Configuration Specification
###################################
" > "${OUTPUT_DIR}"

output_config "Flyte Admin Configuration" flyteadmin
output_config "Flyte Propeller Configuration" flytepropeller
