#!/usr/bin/env bash

set -e

echo "Generating Flyte Configuration Documents"
CUR_DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
ROOT_DIR=${CUR_DIR}/..
OUTPUT_DIR="${ROOT_DIR}"/rsts/deployment/cluster_config
GOBIN=${GOPATH:-~/go}/bin

# The version should same as https://github.com/flyteorg/flyte/blob/6b92b72f508d0603fa44153a4e30cf81be76adfd/script/release.sh#L5-L8
FLYTEADMIN_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flyteadmin/releases/latest" | jq -r .tag_name)
FLYTEPROPELLER_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/flytepropeller/releases/latest" | jq -r .tag_name)
DATACATALOG_TAG=$(curl --silent "https://api.github.com/repos/flyteorg/datacatalog/releases/latest" | jq -r .tag_name)

GO111MODULE=on go get github.com/flyteorg/flyteadmin/cmd@${FLYTEADMIN_TAG}
mv $GOBIN/cmd $GOBIN/flyteadmin
GO111MODULE=on go get github.com/flyteorg/flytepropeller/cmd/controller@${FLYTEPROPELLER_TAG}
mv $GOBIN/controller $GOBIN/flytepropeller
GO111MODULE=on go get github.com/flyteorg/datacatalog/cmd@${DATACATALOG_TAG}
mv $GOBIN/cmd $GOBIN/datacatalog
git clone https://github.com/flyteorg/flyteadmin.git && cd flyteadmin && go build cmd/scheduler/main.go
mv main $GOBIN/scheduler && rm -rf ../flyteadmin

output_config () {
CONFIG_NAME=$1
COMPONENT=$2
COMMAND=$3
OUTPUT_PATH=${OUTPUT_DIR}/${COMMAND}_config.rst

if [ -z "$CONFIG_NAME" ]; then
  log_err "output_config CONFIG_NAME value not specified in arg1"
  return 1
fi

if [ -z "$COMPONENT" ]; then
  log_err "output_config COMPONENT value not specified in arg2"
  return 1
fi

echo ".. _$COMPONENT-config-specification:

#########################################
Flyte $CONFIG_NAME Configuration
#########################################
" > "${OUTPUT_PATH}"

$GOBIN/$COMMAND config docs >> "${OUTPUT_PATH}"
}

output_config "Admin" flyteadmin flyteadmin
output_config "Propeller" flytepropeller flytepropeller
output_config "Datacatalog" flytedatacatalog datacatalog
output_config "Scheduler" flytescheduler scheduler