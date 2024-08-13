#!/usr/bin/env bash

# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'FLYTEORG/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/flyteorg/boilerplate/blob/master/Readme.rst

set -e

echo ""
echo "------------------------------------"
echo "           DOCKER BUILD"
echo "------------------------------------"
echo ""

if [ -n "$REGISTRY" ]; then
  # Do not push if there are unstaged git changes
  CHANGED=$(git status --porcelain)
  if [ -n "$CHANGED" ]; then
    echo "Please commit git changes before pushing to a registry"
    exit 1
  fi
fi


GIT_SHA=$(git rev-parse HEAD)

IMAGE_TAG_SUFFIX=""
# for intermediate build phases, append -$BUILD_PHASE to all image tags
if [ -n "$BUILD_PHASE" ]; then
  IMAGE_TAG_SUFFIX="-${BUILD_PHASE}"
fi

IMAGE_TAG_WITH_SHA="${IMAGE_NAME}:${GIT_SHA}${IMAGE_TAG_SUFFIX}"

RELEASE_SEMVER=$(git describe --tags --exact-match "$GIT_SHA" 2>/dev/null) || true
if [ -n "$RELEASE_SEMVER" ]; then
  IMAGE_TAG_WITH_SEMVER="${IMAGE_NAME}:${RELEASE_SEMVER}${IMAGE_TAG_SUFFIX}"
fi

# build the image
# passing no build phase will build the final image
docker build -t "$IMAGE_TAG_WITH_SHA" --target=${BUILD_PHASE} .
echo "${IMAGE_TAG_WITH_SHA} built locally."

# if REGISTRY specified, push the images to the remote registry
if [ -n "$REGISTRY" ]; then

  if [ -n "${DOCKER_REGISTRY_PASSWORD}" ]; then
    docker login --username="$DOCKER_REGISTRY_USERNAME" --password="$DOCKER_REGISTRY_PASSWORD"
  fi

  docker tag "$IMAGE_TAG_WITH_SHA" "${REGISTRY}/${IMAGE_TAG_WITH_SHA}"

  docker push "${REGISTRY}/${IMAGE_TAG_WITH_SHA}"
  echo "${REGISTRY}/${IMAGE_TAG_WITH_SHA} pushed to remote."

  # If the current commit has a semver tag, also push the images with the semver tag
  if [ -n "$RELEASE_SEMVER" ]; then

    docker tag "$IMAGE_TAG_WITH_SHA" "${REGISTRY}/${IMAGE_TAG_WITH_SEMVER}"

    docker push "${REGISTRY}/${IMAGE_TAG_WITH_SEMVER}"
    echo "${REGISTRY}/${IMAGE_TAG_WITH_SEMVER} pushed to remote."

  fi
fi
