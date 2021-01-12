#!/bin/bash

# This file is meant to be run within the docker container that has the appropriate repos checked out, and mounted
# docker run -t -v ${BASEDIR}:/base -v ${REPOS_DIR}:/repos -v ${BASEDIR}/_rsts:/_rsts lyft/docbuilder:v2.2.0 /base/docs_infra/in_container_rst_generation.sh
#
# It will expect flyteidl and flytekit checked out at the /repos folder.

set -e
set -x

# We're generating documentation for flytesdk, which references flyteidl.  autodoc complains if the python libraries
# referenced are not installed, so let's install them
pip install -U flytekit[all3]==${FLYTEKIT_VERSION}

# Generate the RST files for flytekit
sphinx-apidoc --force --tocfile index --ext-autodoc --output-dir /repos/flytekit/docs /repos/flytekit/flytekit

# Create a temp directory where we can aggregate all the rst files from all locations
mkdir /rsts

# The toctree in this index file requires that the idl/sdk rsts are in the same folder
cp -R /repos/flyteidl/gen/pb-protodoc/flyteidl /rsts/
cp -R /repos/flytekit/docs /rsts/flytekit

cp -R /rsts/* /_rsts
