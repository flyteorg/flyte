#!/bin/bash
set -e

DIR=$(pwd)
rm -rf $DIR/gen

# Override system locale during protos/docs generation to ensure consistent sorting (differences in system locale could e.g. lead to differently ordered docs)
export LC_ALL=C.UTF-8

# Buf migration
docker run -u $(id -u):$(id -g) -e "BUF_CACHE_DIR=/tmp/cache" --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf generate

# Unfortunately the python protoc plugin does not add __init__.py files to the generated code
# (as described in https://github.com/protocolbuffers/protobuf/issues/881). One of the
# suggestions is to manually create such files, which is what we do here:
find gen/pb_python -type d -exec touch {}/__init__.py \;

# Generate JS code
docker run --rm -u $(id -u):$(id -g) -v $DIR:/defs schottra/docker-protobufjs:v0.0.2 --module-name flyteidl -d protos/flyteidl/core -d protos/flyteidl/event -d protos/flyteidl/admin -d protos/flyteidl/service -- --root flyteidl -t static-module -w default --no-delimited --force-long --no-convert -p /defs/protos

# This section is used by Travis CI to ensure that the generation step was run
if [ -n "$DELTA_CHECK" ]; then
	DIRTY=$(git status --porcelain)
	if [ -n "$DIRTY" ]; then
		echo "FAILED: Protos updated without committing generated code."
		echo "Ensure make generate has run and all changes are committed."
		DIFF=$(git diff)
		echo "diff detected: $DIFF"
		DIFF=$(git diff --name-only)
		echo "files different: $DIFF"
		exit 1
	else
		echo "SUCCESS: Generated code is up to date."
	fi
fi
