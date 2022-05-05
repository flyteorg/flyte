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

rm -rf flytearchives
mkdir flytearchives
cd flytearchives
curl https://github.com/flyteorg/flyteadmin/releases/download/${FLYTEADMIN_TAG}/flyteadmin_${FLYTEADMIN_TAG}_linux_x86_64.tar.gz --output flyteadmin.tar.gz -s -L && tar -xvf flyteadmin.tar.gz
mv flyteadmin $GOBIN/flyteadmin
mv flytescheduler $GOBIN/flytescheduler
curl https://github.com/flyteorg/flytepropeller/releases/download/${FLYTEPROPELLER_TAG}/flytepropeller_${FLYTEPROPELLER_TAG}_linux_x86_64.tar.gz --output flytepropeller.tar.gz -s -L && tar -xvf flytepropeller.tar.gz
mv flytepropeller $GOBIN/flytepropeller
curl https://github.com/flyteorg/datacatalog/releases/download/${DATACATALOG_TAG}/datacatalog_${DATACATALOG_TAG}_linux_x86_64.tar.gz --output datacatalog.tar.gz -s -L && tar -xvf datacatalog.tar.gz
mv datacatalog $GOBIN/datacatalog
# Clean up tar archives
cd ../../ && rm -rf flytearchives

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
output_config "Scheduler" flytescheduler flytescheduler