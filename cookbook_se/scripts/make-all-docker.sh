#!/bin/bash

echo ${REGISTRY}
DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" && pwd )"

for i in `ls ${ENV_BASE_PATH}`;do
  if [[ "${i}" != "common" ]]; then
    ENV_PATH=${ENV_BASE_PATH}${i}
    cd ${ENV_PATH};
    echo "Running make command in `pwd`"
    make ${1};
    cd ${DIR};
  fi
done
