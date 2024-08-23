export REPOSITORY=flytectl
export REPO_ROOT=..
include ../boilerplate/flyte/docker_build/Makefile
include ../boilerplate/flyte/golang_test_targets/Makefile
include ../boilerplate/flyte/end2end/Makefile

GIT_VERSION := $(shell git describe --dirty --tags --long --match 'flytectl/*' --first-parent | sed 's/^flytectl\///')
GIT_HASH := $(shell git rev-parse --short HEAD)
TIMESTAMP := $(shell date '+%Y-%m-%d')
PACKAGE ?=github.com/flyteorg/flyte/flytestdlib

LD_FLAGS="-s -w -X $(PACKAGE)/version.Version=$(GIT_VERSION) -X $(PACKAGE)/version.Build=$(GIT_HASH) -X $(PACKAGE)/version.BuildTime=$(TIMESTAMP)"

define PIP_COMPILE
pip-compile $(1) --upgrade --verbose
endef

compile:
	go build -o bin/flytectl -ldflags=$(LD_FLAGS) main.go

compile_debug:
	go build -gcflags='all=-N -l' -o bin/flytectl main.go

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

.PHONY: install-piptools
install-piptools:
	pip install -U pip-tools

.PHONY: test_unit_without_flag
test_unit_without_flag:
	go test ./... -race -coverprofile=coverage.temp.txt -covermode=atomic
	cat coverage.temp.txt  | grep -v "_flags.go" > coverage.txt
	rm coverage.temp.txt
	curl -s https://codecov.io/bash > codecov_bash.sh && bash codecov_bash.sh
