export REPOSITORY=flytectl
include boilerplate/lyft/golang_test_targets/Makefile

define PIP_COMPILE
pip-compile $(1) --upgrade --verbose
endef

generate:
	go test github.com/flyteorg/flytectl/cmd --update

compile:
	go build -o bin/flytectl main.go

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: install-piptools
install-piptools:
	pip install -U pip-tools

.PHONY: doc-requirements.txt
doc-requirements.txt: doc-requirements.in install-piptools
	$(call PIP_COMPILE,doc-requirements.in)
