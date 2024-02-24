#!/bin/bash

export REPOSITORY=flyteidl
include ../boilerplate/flyte/golang_test_targets/Makefile

define PIP_COMPILE
pip-compile $(1) --upgrade --verbose
endef

.PHONY: generate
generate: install # install tools, generate protos, mock, pflags and  get doc dependencies
	./generate_protos.sh
	./generate_mocks.sh
	go generate ./...

.PHONY: test
test: install # ensures generate_protos script has been run
	git diff
	./generate_mocks.sh
	go generate ./...
	DELTA_CHECK=true ./generate_protos.sh

.PHONY: build_python
build_python:
	@python setup.py sdist

.PHONY: install-piptools
install-piptools:
	pip install -U pip-tools

.PHONY: doc-requirements.txt
doc-requirements.txt: doc-requirements.in install-piptools
	$(call PIP_COMPILE,doc-requirements.in)

PLACEHOLDER_NPM := \"version\": \"0.0.0-develop\"

.PHONY: update_npmversion
update_npmversion:
	grep "$(PLACEHOLDER_NPM)" "package.json"
	sed -i "s/$(PLACEHOLDER_NPM)/\"version\":  \"${VERSION}\"/g" "package.json"
