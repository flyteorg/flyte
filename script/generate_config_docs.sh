#!/usr/bin/env bash

set -e

echo "Generating Flyte Configuration Documents"
CUR_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" >/dev/null && pwd)"
ROOT_DIR=${CUR_DIR}/..
OUTPUT_DIR="${ROOT_DIR}"/docs/deployment/configuration/generated
GOBIN=${GOPATH:-~/go}/bin

make -C datacatalog compile
mv datacatalog/bin/datacatalog ${GOBIN}/datacatalog
make -C flyteadmin compile
mv flyteadmin/bin/flyteadmin ${GOBIN}/flyteadmin
make -C flyteadmin compile_scheduler
mv flyteadmin/bin/flytescheduler ${GOBIN}/scheduler
make -C flytepropeller compile_flytepropeller
mv flytepropeller/bin/flytepropeller ${GOBIN}/flytepropeller

# Config files are needed to generate docs, so we generate an empty
# file and reuse it to invoke the docs command in all components.
EMPTY_CONFIG_FILE=empty-config.yaml
touch empty-config.yaml

output_config() {
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
" >"${OUTPUT_PATH}"

  $GOBIN/$COMMAND config --config $EMPTY_CONFIG_FILE docs >>"${OUTPUT_PATH}"
}

output_config "Admin" flyteadmin flyteadmin
output_config "Propeller" flytepropeller flytepropeller
output_config "Datacatalog" flytedatacatalog datacatalog
output_config "Scheduler" flytescheduler scheduler
