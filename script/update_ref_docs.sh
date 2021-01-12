#!/bin/bash

# Note that this file is meant to be run on OSX by a user with the necessary GitHub privileges.
# This script
#    a) clones the two Flyte repositories from which additional RSTs not in this flyte repo, need to be generated.
#       namely flytekit, and flyteidl
#    b) runs a docker image to parse through the cloned repos, and creates the RSTs in the _rsts/ folder, which has
#       been added to gitignore.

set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
BASEDIR="${DIR}/.."

# Set up a temp directory
mkdir ${BASEDIR}/_repos || true
REPOS_DIR=`mktemp -d "${BASEDIR}/_repos/XXXXXXXXX"`

# Clone all repos
echo "Cloning Flyteidl"
git clone https://github.com/lyft/flyteidl.git --single-branch --branch v${FLYTEIDL_VERSION} ${REPOS_DIR}/flyteidl
echo "Cloning Flytekit"
git clone https://github.com/lyft/flytekit.git --single-branch --branch v${FLYTEKIT_VERSION} ${REPOS_DIR}/flytekit

# Generate documentation by running script inside the generation container
docker run --rm -t -e FLYTEKIT_VERSION=${FLYTEKIT_VERSION} -v ${BASEDIR}:/base -v ${REPOS_DIR}:/repos -v ${BASEDIR}/_rsts:/_rsts ghcr.io/nuclyde-io/docbuilder:e461362c9da2415ac5419e4b2b0f13f839bdd1fe /base/docs_infra/in_container_rst_generation.sh

# Cleanup
rm -rf ${REPOS_DIR}/* || true
