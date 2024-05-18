#!/usr/bin/env bash

set -eo pipefail

# Available environment variables
# - DEV_DOCS_WATCH: If set, the docs will be built and served using sphinx-autobuild for live updates
# - FLYTEKIT_LOCAL_PATH: If set, the local path to flytekit will be used instead of the source code from the flyteorg/flytekit repo
# - FLYTECTL_LOCAL_PATH: If set, the local path to flytectl will be used instead of the source code from the flyteorg/flytectl repo
# - FLYTESNACKS_LOCAL_PATH: If set, the local path to flytesnacks will be used instead of the source code from the flyteorg/flytesnacks repo
#
# Example usages:
#   ./script/local_build_docs.sh
#   DEV_DOCS_WATCH=1 ./script/local_build_docs.sh
#   FLYTEKIT_LOCAL_PATH=$(realpath ../flytekit) ./script/local_build_docs.sh

DOCKER_ENV=()
DOCKER_VOLUME_MAPPING=("-v" ".:/flyte")
DOCKER_PORT_MAPPING=()
BUILD_CMD=("sphinx-build" "-M" "html" "." "_build" "-W")

if [ -n "$DEV_DOCS_WATCH" ]; then
  DOCKER_PORT_MAPPING+=("-p" "8000:8000")
  BUILD_CMD=("sphinx-autobuild" "--host" "0.0.0.0" "--re-ignore" "'_build|_src|_tags|api|examples|flytectl|flytekit|flytesnacks|protos|tests'" "." "_build/html")
fi

if [ -n "$FLYTEKIT_LOCAL_PATH" ]; then
  DOCKER_VOLUME_MAPPING+=("-v" "$FLYTEKIT_LOCAL_PATH:/flytekit")
  DOCKER_ENV+=("--env" "FLYTEKIT_LOCAL_PATH=/flytekit")
fi

if [ -n "$FLYTECTL_LOCAL_PATH" ]; then
  DOCKER_VOLUME_MAPPING+=("-v" "$FLYTECTL_LOCAL_PATH:/flytectl")
  DOCKER_ENV+=("--env" "FLYTECTL_LOCAL_PATH=/flytectl")
fi

if [ -n "$FLYTESNACKS_LOCAL_PATH" ]; then
  DOCKER_VOLUME_MAPPING+=("-v" "$FLYTESNACKS_LOCAL_PATH:/flytesnacks")
  DOCKER_ENV+=("--env" "FLYTESNACKS_LOCAL_PATH=/flytesnacks")
fi

DOCKER_CMD=("docker" "run" "--platform" "linux/amd64" "--rm" "--pull" "never" "${DOCKER_ENV[@]}" "${DOCKER_PORT_MAPPING[@]}" "${DOCKER_VOLUME_MAPPING[@]}" "flyte-dev-docs:latest" "${BUILD_CMD[@]}")

echo "${DOCKER_CMD[*]}"

"${DOCKER_CMD[@]}"
