#!/bin/bash

export REPOSITORY=flyteidl
include boilerplate/lyft/golang_test_targets/Makefile

define PIP_COMPILE
pip-compile $(1) --upgrade --verbose
endef

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: generate
generate: install doc_gen_deps # install tools, generate protos, doc dependecies, mocks and pflags
	./generate_protos.sh
	./generate_mocks.sh
	go generate ./...

.PHONY: test
test: install # ensures generate_protos script has been run
	git diff
	./generate_mocks.sh
	go generate ./...
	DELTA_CHECK=true ./generate_protos.sh

.PHONY: test_unit
test_unit:
	# we cannot use test_unit from go.mk because generated files contain commented import statements that
	# go tries to intepret. So we need to use go list to get the packages that go understands.
	go test -cover `go list ./...` -race

.PHONY: build_python
build_python:
	@python setup.py sdist

.PHONY: install-piptools
install-piptools:
	pip install -U pip-tools

.PHONY: doc_gen_deps # these dependencies are required by protoc gen doc for the protos which have external library dependencies. 
# which includes grpc-gateway, googleapis, k8.io/api and apimachinery, protocolbuffers 
doc_gen_deps:
	rm -rf tmp/doc_gen_deps tmp/protocolbuffers tmp/grpc-gateway tmp/k8s.io
	git clone --depth 1 https://github.com/googleapis/googleapis tmp/doc_gen_deps
	rm -rf tmp/doc_gen_deps/.git
	git clone --depth 1 https://github.com/protocolbuffers/protobuf tmp/protocolbuffers
	cp -r tmp/protocolbuffers/src/* tmp/doc_gen_deps/
	rm -rf tmp/protocolbuffers
	git -c advice.detachedHead=false clone --depth 1 --branch v1.15.2 https://github.com/grpc-ecosystem/grpc-gateway tmp/grpc-gateway #v1.15.2 is used to keep the grpc-gateway version in sync with generated protos which is using the LYFT image
	cp -r tmp/grpc-gateway/protoc-gen-swagger  tmp/doc_gen_deps/
	rm -rf tmp/grpc-gateway
	git clone --depth 1 https://github.com/kubernetes/api tmp/k8s.io/api
	git clone --depth 1 https://github.com/kubernetes/apimachinery tmp/k8s.io/apimachinery
	cp -r tmp/k8s.io tmp/doc_gen_deps/
	rm -rf tmp/k8s.io

.PHONY: doc-requirements.txt
doc-requirements.txt: doc-requirements.in install-piptools
	$(call PIP_COMPILE,doc-requirements.in)

PLACEHOLDER := "__version__\ =\ \"develop\""
PLACEHOLDER_NPM := \"version\": \"develop\"

.PHONY: update_pyversion
update_pyversion:
	grep "$(PLACEHOLDER)" "setup.py"
	sed -i "s/$(PLACEHOLDER)/__version__ = \"${VERSION}\"/g" "setup.py"

.PHONY: update_npmversion
update_npmversion:
	grep "$(PLACEHOLDER_NPM)" "package.json"
	sed -i "s/$(PLACEHOLDER_NPM)/\"version\":  \"${VERSION}\"/g" "package.json"
