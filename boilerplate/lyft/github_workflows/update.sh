#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
#
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

set -e

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"

mkdir -p ${DIR}/../../../.github/workflows

echo "     - generating github action workflows in root directory."
sed -e "s/{{REPOSITORY}}/${REPOSITORY}/g" ${DIR}/master.yml > ${DIR}/../../../.github/workflows/master.yml
sed -e "s/{{REPOSITORY}}/${REPOSITORY}/g" ${DIR}/pull_request.yml > ${DIR}/../../../.github/workflows/pull_request.yml
