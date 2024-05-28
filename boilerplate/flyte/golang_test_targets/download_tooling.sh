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
	"github.com/vektra/mockery/v2@v2.40.3"
	"github.com/golangci/golangci-lint/cmd/golangci-lint"
	"github.com/daixiang0/gci"
	"github.com/alvaroloes/enumer"
	"github.com/pseudomuto/protoc-gen-doc/cmd/protoc-gen-doc"
)

# This ensures pflags are up to date.
make -C $REPO_ROOT/flytestdlib compile
cp $REPO_ROOT/flytestdlib/bin/pflags $(go env GOPATH)/bin

tmp_dir=$(mktemp -d -t gotooling-XXX)
echo "Using temp directory ${tmp_dir}"
cp -R ../boilerplate/flyte/golang_support_tools/* $tmp_dir
pushd "$tmp_dir"

for tool in "${tools[@]}"; do
	echo "Installing ${tool}"
	GO111MODULE=on go install $tool
	# If tool is our mockery fork, we need to rename the binary to mockery-fork
	if [[ $tool == "github.com/EngHabu/mockery/cmd/mockery" ]]; then
		echo "Renaming mockery to mockery-fork"
		mv $(go env GOPATH)/bin/mockery $(go env GOPATH)/bin/mockery-fork
	fi
	# If tool is named vektra/mockery/v2, we need to rename the binary to mockery-v2
	if [[ $tool == "github.com/vektra/mockery/v2@v2.40.3" ]]; then
		echo "Renaming mockery to mockery-v2"
		mv $(go env GOPATH)/bin/mockery $(go env GOPATH)/bin/mockery-v2
	fi
done

# Rename the mockery-fork binary to mockery to maintain compatibility with the existing uses
if [ -f $(go env GOPATH)/bin/mockery-fork ]; then
	echo "Renaming mockery-fork to mockery"
	mv $(go env GOPATH)/bin/mockery-fork $(go env GOPATH)/bin/mockery
fi

popd
