#!/bin/bash


set -ex

DIR="$( cd "$( dirname "${BASH_SOURCE[0]}" )" >/dev/null && pwd )"
BASEDIR="${DIR}/.."

# Clean out all old docs, generate everything from scratch every time
rm -rf ${BASEDIR}/docs

# Set up a temp directory
rm -rf ${BASEDIR}/rsts_tmp || true
mkdir ${BASEDIR}/rsts_tmp || true
RSTS_DIR=`mktemp -d "${BASEDIR}/rsts_tmp/XXXXXXXXX"`

# Copy all rst files to the same place
cp -R docs/* ${RSTS_DIR}
cp -R docs/* ${RSTS_DIR}

# Generate documentation by running script inside the generation container
docker run --rm -t -e FLYTEKIT_VERSION=${FLYTEKIT_VERSION} -v ${BASEDIR}:/base -v ${BASEDIR}/docs:/docs -v ${RSTS_DIR}:/rsts ghcr.io/nuclyde-io/docbuilder:e70cfafe3397c3b23d11183973c0f44e0024f025 /base/docs_infra/in_container_html_generation.sh

# Cleanup
rm -rf ${RSTS_DIR} || true
