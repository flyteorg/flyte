#!/bin/bash

# Everything in this file needs to be installed outside of current module
# The reason we cannot turn off module entirely and install is that we need the replace statement in go.mod
# because we are installing a mockery fork. Turning it off would result installing the original not the fork.
# We also want to version all the other tools.  We also want to be able to run go mod tidy without removing the version
# pins.  To facilitate this, we're maintaining two sets of go.mod/sum files - the second one only for tooling.  This is
# the same approach that go 1.14 will take as well.
# See:
#   https://github.com/flyteorg/flyte/issues/129
#   https://github.com/golang/go/issues/30515 for some background context
#   https://github.com/go-modules-by-example/index/blob/5ec250b4b78114a55001bd7c9cb88f6e07270ea5/010_tools/README.md

set -e

# List of tools to go get
# In the format of "<cli>:<package>" or ":<package>" if no cli
tools=(
  "github.com/EngHabu/mockery/cmd/mockery"
  "github.com/flyteorg/flytestdlib/cli/pflags@latest"
  "github.com/golangci/golangci-lint/cmd/golangci-lint@latest"
  "github.com/alvaroloes/enumer"
  "github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc"
)

tmp_dir=$(mktemp -d -t gotooling-XXX)
echo "Using temp directory ${tmp_dir}"
cp -R boilerplate/flyte/golang_support_tools/* $tmp_dir
pushd "$tmp_dir"

for tool in "${tools[@]}"
do
    echo "Installing ${tool}"
    GO111MODULE=on go install $tool
done

popd
