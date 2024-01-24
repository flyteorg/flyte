#!/usr/bin/env bash

set -eu

if [ -z "$1" ]
  then
    echo "Missing tag or sha to pull, i.e., v1.10.7"
    exit 1
fi

REMOTE_REF=$1

ret=0
git remote show upstream > /dev/null 2>&1 || ret=$?
if [[ $ret != 0 ]]; then
    echo "Adding upstream remote"
    git remote add upstream https://github.com/flyteorg/flyte.git
fi

git fetch upstream

ret=0
git show-ref -q "${REMOTE_REF}" || ret=$?
if [[ $ret != 0 ]]; then
    echo "Couldn't find ${REMOTE_REF} in upstream. Exiting."
    exit 1
fi

git checkout -b "flyte-${REMOTE_REF}"
git pull upstream "${REMOTE_REF}"

echo "Successfully pulled ${REMOTE_REF} from upstream"
