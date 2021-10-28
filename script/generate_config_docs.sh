#!/usr/bin/env bash

set -e

echo "Generating Flyte Configuration Documents"
CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
ROOT_DIR=${CUR_DIR}/..
OUTPUT_DIR="${ROOT_DIR}"/rsts/deployment/cluster_config

# GO111MODULE=on go get github.com/flyteorg/flyteadmin/cmd@config_docs

output_config () {
OUTPUT_PATH="${OUTPUT_DIR}"/$2_config.rst

echo ".. _$2-config-specification:

#########################################
Flyte $1 Configuration
#########################################
" > "${OUTPUT_PATH}"

"${GOPATH:-~/go}"/bin/$2 config docs >> "${OUTPUT_PATH}"
}

output_config "Admin" flyteadmin
output_config "Propeller" flytepropeller
