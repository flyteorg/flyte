#!/bin/bash

# Our SDK entrypoint can be configured to call this script
# This script maps that command to the conventional location of the virtual environment in Flyte containers

set -e

. ${VENV}/bin/activate

exec $*
