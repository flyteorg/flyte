#!/bin/bash

# Note that this file is meant to be run on OSX by a user with the necessary GitHub privileges.
# This script
#    a) clones the two Flyte repositories from which additional RSTs not in this flyte repo, need to be generated.
#       namely flytekit, and flyteidl
#    b) runs a docker image to parse through the cloned repos, and creates the RSTs in the _rsts/ folder, which has
#       been added to gitignore.

set -ex

# Set up a temp directory
mkdir -p _repos

# Clone all repos
echo "Cloning Flyteidl"
git clone https://github.com/lyft/flyteidl.git _repos/flyteidl
echo "Cloning Flytekit"
git clone https://github.com/lyft/flytekit.git _repos/flytekit
