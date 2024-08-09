#!/bin/bash
docker run --rm -u $(id -u):$(id -g) -e "BUF_CACHE_DIR=/tmp/cache" --volume "$(pwd):/workspace" --workdir /workspace bufbuild/buf generate
cd worker && cargo fmt --all
