.SILENT:

include common/Makefile
include common/parent.mk

define PIP_COMPILE
pip-compile $(1) ${PIP_ARGS} --upgrade --verbose
endef

dev-requirements.txt: export CUSTOM_COMPILE_COMMAND := $(MAKE) dev-requirements.txt
dev-requirements.txt: dev-requirements.in install-piptools
	$(call PIP_COMPILE,dev-requirements.in)

.PHONY: dev-requirements
dev-requirements: dev-requirements.txt

docs-requirements.txt: export CUSTOM_COMPILE_COMMAND := $(MAKE) docs-requirements.txt
docs-requirements.txt: docs-requirements.in install-piptools
	$(call PIP_COMPILE,docs-requirements.in)

.PHONY: docs-requirements
docs-requirements: docs-requirements.txt

.PHONY: enter_sandbox
enter_sandbox: docker_build
	docker run -e PROJECT=${PROJECT} -v `pwd`:/root -it ${FULL_IMAGE_NAME}:${VERSION} bash

.PHONE: run-all-examples
run-examples: ## Runs all examples
	sh scripts/run-all-examples.sh

.PHONY: archive
archive: # archive directories (e.g. _pb_output/)
	bash scripts/ci_helper.sh RELEASE

.PHONY: register-workflow-ci
register-workflow-ci:
	bash scripts/ci_helper.sh REGISTER
