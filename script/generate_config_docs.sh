#!/usr/bin/env bash

set -e

echo "Generating Flyte Configuration Documents"
CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
ROOT_DIR=${CUR_DIR}/..
OUTPUT_DIR="${ROOT_DIR}"/rsts/deployment/cluster_config
GOBIN=${GOPATH:-~/go}/bin

FLYTEADMIN_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteadmin/releases/latest" | jq -r .tag_name)
FLYTEPROPELLER_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flytepropeller/releases/latest" | jq -r .tag_name)

GO111MODULE=on go get github.com/flyteorg/flyteadmin/cmd@${FLYTEADMIN_TAG}
GO111MODULE=on go get github.com/flyteorg/flytepropeller/cmd/controller@${FLYTEPROPELLER_TAG}
mv $GOBIN/cmd $GOBIN/flyteadmin
mv $GOBIN/controller $GOBIN/flytepropeller

output_config () {
OUTPUT_PATH="${OUTPUT_DIR}"/$2_config.rst

echo ".. _$2-config-specification:

#########################################
Flyte $1 Configuration
#########################################
" > "${OUTPUT_PATH}"

$GOBIN/$2 config docs >> "${OUTPUT_PATH}"
}

output_config "Admin" flyteadmin
output_config "Propeller" flytepropeller
