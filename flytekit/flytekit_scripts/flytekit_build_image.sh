#!/usr/bin/env bash
set -e

# NB: This script is bundled with the FlyteKit SDK and is the recommended way for users to build their images.
# This build script ensures that the resulting workflow image that the Flyte ecosystem executes is built
# in a standardized way, complete with all the environment variables that the rest of the system will rely on.

if [ -z "$1" ]; then
    echo usage: ./flytekit_build_image.sh [/path/to/Dockerfile or /path/to/repo] [optional tag prefix]
    exit
fi

echo ""
echo "------------------------------------"
echo "           DOCKER BUILD"
echo "------------------------------------"
echo ""

DIRPATH=""
DOCKERFILE_PATH=""
if [ -d "$1" ]; then
  DIRPATH=$(cd "$1" && pwd)
  DOCKERFILE_PATH=${DIRPATH}/"Dockerfile"
elif [ -f "$1" ]; then
  DIRPATH=$(cd "$(dirname "$1")" && pwd)
  DOCKERFILE_PATH=${DIRPATH}/$(basename "$1")
else
  exit 1
fi

PREFIX="$2-"
if [ -z "$2" ]; then
  PREFIX=""
fi

# Go into the directory representing the user's repo
pushd "${DIRPATH}"

# Grab the repo name from the argument if not already defined
# Note that this repo name will be the name of the Docker image.
if [ -z "${IMAGE_NAME}" ]; then
  IMAGE_NAME=${PWD##*/}
fi

# Do not do anything if there are unstaged git changes
CHANGED=$(git status --porcelain)
if [ -n "$CHANGED" ]; then
  echo "Please commit git changes before building"
  exit 1
fi

# set default tag to the latest git SHA
if [ -z "$TAG" ]; then
  TAG=$(git rev-parse HEAD)
fi

# If the registry does not exist, don't worry about it, and let users build locally
# This is the name of the image that will be injected into the container as an environment variable
if [ -n "$REGISTRY" ]; then
  FLYTE_INTERNAL_IMAGE=${REGISTRY}/${IMAGE_NAME}:${PREFIX}${TAG}
else
  FLYTE_INTERNAL_IMAGE=${IMAGE_NAME}:${PREFIX}${TAG}
fi

DOCKER_PLATFORM_OPT=()
# Check if the user set the target build architecture, if not use the default instead.
if [ -n "$TARGET_PLATFORM_BUILD" ]; then
  DOCKER_PLATFORM_OPT=(--platform "$TARGET_PLATFORM_BUILD")
else
  TARGET_PLATFORM_BUILD="default"
fi

echo "Building: $FLYTE_INTERNAL_IMAGE for $TARGET_PLATFORM_BUILD architecture"

# This build command is the raison d'etre of this script, it ensures that the version is injected into the image itself
docker build . "${DOCKER_PLATFORM_OPT[@]}" --build-arg tag="${FLYTE_INTERNAL_IMAGE}" -t "${FLYTE_INTERNAL_IMAGE}" -f "${DOCKERFILE_PATH}"

echo "$IMAGE_NAME built locally."

# Create the appropriate tags
echo "${FLYTE_INTERNAL_IMAGE} tagged"
if [ -n "$REGISTRY" ]; then
  # Also push if there's a registry to push to
  if [[ "${REGISTRY}" == "docker.io"*  && -z "${NOPUSH}" ]]; then
    docker login --username="${DOCKERHUB_USERNAME}" --password="${DOCKERHUB_PASSWORD}"
  fi

  if [[ -z "${NOPUSH}" ]]; then
    docker push "${FLYTE_INTERNAL_IMAGE}"
    echo "${FLYTE_INTERNAL_IMAGE} pushed to remote"
  fi
fi

popd

echo ""
echo "------------------------------------"
echo "              SUCCESS"
echo "------------------------------------"
echo ""
