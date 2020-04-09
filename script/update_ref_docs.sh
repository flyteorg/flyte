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

function clone_repos()
{
    REPOS=`cat ${DIR}/dependent_repos.txt`
    for REPO in ${REPOS}; do
      git clone https://${REPO}.git ${REPOS_DIR}/flytekit
    done
}

# Clone all repos
$(clone_repos)

# Generate documentation by running script inside the generation container
docker run -t -v ${BASEDIR}:/base -v ${REPOS_DIR}:/repos -v ${BASEDIR}/_rsts:/_rsts lyft/docbuilder:v2.2.0 /base/docs_infra/in_container_rst_generation.sh

# Cleanup
rm -rf ${REPOS_DIR}/* || true
