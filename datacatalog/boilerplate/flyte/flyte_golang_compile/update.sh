#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'FLYTEORG/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/flyteorg/boilerplate/blob/master/Readme.rst

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

echo "     - generating ${DIR}/flyte_golang_compile.sh"
sed -e "s/{{REPOSITORY}}/${REPOSITORY}/g" ${DIR}/flyte_golang_compile.Template > ${DIR}/flyte_golang_compile.sh
