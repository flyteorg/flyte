#!/usr/bin/env bash
set -e

# NB: This script is originally bundled with the FlyteKit SDK with slight modifications

echo ""
echo "------------------------------------"
echo "           DOCKER BUILD"
echo "------------------------------------"
echo ""

# Go into the directory representing the user's repo
pushd "$1"

if [ -n "$REGISTRY" ]; then
  # Do not push if there are unstaged git changes
  CHANGED=$(git status --porcelain)
  if [ -n "$CHANGED" ]; then
    echo "Please commit git changes before pushing to a registry"
    exit 1
  fi
fi

# set default tag to the latest git SHA
if [ -z "$TAG" ]; then
  TAG=$(git rev-parse HEAD)
fi

# If the registry does not exist, don't worry about it, and let users build locally
# This is the name of the image that will be injected into the container as an environment variable
if [ -n "$REGISTRY" ]; then
  FLYTE_INTERNAL_IMAGE=${REGISTRY}/${IMAGE_NAME}:${TAG}
else
  FLYTE_INTERNAL_IMAGE=${IMAGE_NAME}:${TAG}
fi
echo "Building: $FLYTE_INTERNAL_IMAGE"

# This build command is the raison d'etre of this script, it ensures that the version is injected into the image itself
docker build . --build-arg tag="$FLYTE_INTERNAL_IMAGE" -t "$IMAGE_NAME" -f ./Dockerfile
echo "$IMAGE_NAME built locally."

# Create the appropriate tags
docker tag "${IMAGE_NAME}:latest" "${IMAGE_NAME}:${TAG}"
echo "${IMAGE_NAME}:latest also tagged with ${IMAGE_NAME}:${TAG}"
if [ -n "$REGISTRY" ]; then
  docker tag "${IMAGE_NAME}:latest" "${FLYTE_INTERNAL_IMAGE}"
  echo "${IMAGE_NAME}:latest also tagged with ${FLYTE_INTERNAL_IMAGE}"

  # Also push if there's a registry to push to
  if [[ "${REGISTRY}" == "docker.io"* ]]; then
    docker login --username="${DOCKERHUB_USERNAME}" --password="${DOCKERHUB_PASSWORD}"
  fi
  docker push "${FLYTE_INTERNAL_IMAGE}"
  echo "${FLYTE_INTERNAL_IMAGE} pushed to remote"
fi

popd

echo ""
echo "------------------------------------"
echo "              SUCCESS"
echo "------------------------------------"
echo ""
