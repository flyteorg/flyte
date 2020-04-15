#!/bin/bash

# This file is meant to be run within the docker container, and mounted
# /rsts as the inputs dir
# /docs as the output dir
# /base as the dir where this script lives
# docker run -t -v ${BASEDIR}:/base -v ${BASEDIR}/docs:/docs -v ${BASEDIR}/rsts_tmp:/rsts lyft/docbuilder:v2.2.0 /base/docs_infra/in_container_html_generation.sh
#
# It will expect flyteidl and flytekit checked out at the /repos folder.

set -e
set -x

# We're generating documentation for flytekit, which references flyteidl.  autodoc complains if the python libraries
# referenced are not installed, so let's install them
echo "Installing flytekit==${FLYTEKIT_VERSION}"
pip install -U flytekit[all]==${FLYTEKIT_VERSION}

sphinx-build -Q -b html -c /rsts /rsts /docs/
