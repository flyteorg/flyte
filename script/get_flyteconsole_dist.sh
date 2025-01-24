#!/usr/bin/env bash

set -ex

ARCH="$(uname -m)"
case ${ARCH} in
x86_64|amd64)
  IMAGE_ARCH=amd64
  ;;
aarch64|arm64)
  IMAGE_ARCH=arm64
  ;;
*)
  >&2 echo "ERROR: Unsupported architecture: ${ARCH}"
  exit 1
  ;;
esac

FLYTECONSOLE_IMAGE="ghcr.io/flyteorg/flyteconsole:${FLYTECONSOLE_VERSION:-latest}"
IMAGE_DIGEST="$(docker manifest inspect --verbose ${FLYTECONSOLE_IMAGE} | \
    jq --arg IMAGE_ARCH "${IMAGE_ARCH}" --raw-output \
      '.[] | select(.Descriptor.platform.architecture == $IMAGE_ARCH) | .Descriptor.digest')"

# Short circuit if we already have the correct distribution
[ -f cmd/single/dist/.digest ] && grep -Fxq ${IMAGE_DIGEST} cmd/single/dist/.digest && exit 0

# Create container from desired image
CONTAINER_ID=$(docker create ghcr.io/flyteorg/flyteconsole:${FLYTECONSOLE_VERSION:-latest})
trap 'docker rm -f ${CONTAINER_ID}' EXIT

# Copy distribution
rm -rf cmd/single/dist
docker cp ${CONTAINER_ID}:/app cmd/single/dist
printf '%q' ${IMAGE_DIGEST} > cmd/single/dist/.digest
