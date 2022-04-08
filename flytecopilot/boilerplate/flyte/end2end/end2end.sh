#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'FLYTEORG/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/flyteorg/boilerplate/blob/master/Readme.rst
set -e

CONFIG_FILE=$1; shift
EXTRA_FLAGS=( "$@" )

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

LATEST_VERSION=$(curl --silent "https://api.github.com/repos/flyteorg/flytesnacks/releases/latest" | jq -r .tag_name)

python ./boilerplate/flyte/end2end/run-tests.py $LATEST_VERSION P0,P1 $CONFIG_FILE ${EXTRA_FLAGS[@]}
