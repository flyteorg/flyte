# This is used by the image building script referenced below. Normally it just takes the directory name but in this
# case we want it to be called something else.
IMAGE_NAME=flytecookbook_se
VERSION=$(shell ./version.sh)

define PIP_COMPILE
pip-compile $(1) --upgrade --verbose
endef

# If the REGISTRY environment variable has been set, that means the image name will not just be tagged as
#   flytecookbook:<sha> but rather,
#   docker.io/lyft/flytecookbook:<sha> or whatever your REGISTRY is.
ifneq ($(origin REGISTRY), undefined)
	FULL_IMAGE_NAME = ${REGISTRY}/${IMAGE_NAME}
else
	FULL_IMAGE_NAME = ${IMAGE_NAME}
endif

# The Flyte project that we want to register under
PROJECT=flytesnacks

.SILENT: help
.PHONY: help
help:
	echo Available recipes:
	cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' | awk 'BEGIN { FS = ":.*?## " } { cnt++; a[cnt] = $$1; b[cnt] = $$2; if (length($$1) > max) max = length($$1) } END { for (i = 1; i <= cnt; i++) printf "  $(shell tput setaf 6)%-*s$(shell tput setaf 0) %s\n", max, a[i], b[i] }'
	tput sgr0

.PHONY: debug
debug:
	echo "IMAGE NAME ${IMAGE_NAME}"
	echo "FULL IMAGE NAME ${FULL_IMAGE_NAME}"
	echo "VERSION TAG ${VERSION}"
	echo "REGISTRY ${REGISTRY}"

.PHONY: docker_build
docker_build:
	NOPUSH=1 IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh ./Dockerfile
	NOPUSH=1 IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh ./Dockerfile.spark spark

# This should only be used by Admins to push images to the public Dockerhub repo. Make sure you
# specify REGISTRY=docker.io/lyft before the make command otherwise this won't actually push
.PHONY: docker_push
docker_push:
	IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh ./Dockerfile
	IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh ./Dockerfile.spark spark

.PHONY: in_container_register_sandbox
in_container_register_sandbox:
	pyflyte -p ${PROJECT} -d development --config /root/sandbox.config register workflows

.PHONY: register_sandbox
register_sandbox: docker_build
	docker run --rm --network=host ${FULL_IMAGE_NAME}:${VERSION} /usr/local/bin/flytekit_venv make in_container_register_sandbox

.PHONY: in_container_serialize_sandbox
in_container_serialize_sandbox:
	pyflyte -p ${PROJECT} -d development --config /root/sandbox.config serialize workflows -f /tmp/output

.PHONY: serialize_sandbox
serialize_sandbox: docker_build
	echo ${CURDIR}
	mkdir ${CURDIR}/_pb_output || true
	docker run -v ${CURDIR}/_pb_output:/tmp/output ${FULL_IMAGE_NAME}:${VERSION} make in_container_serialize_sandbox

.PHONY: enter_sandbox
enter_sandbox: docker_build
	docker run -e PROJECT=${PROJECT} -v `pwd`:/root -it ${FULL_IMAGE_NAME}:${VERSION} bash

.PHONY: fmt
fmt: ## Format code with black and isort
	black .
	isort .

.PHONY: install-piptools
install-piptools:
	pip install -U pip-tools

.PHONY: setup
setup: install-piptools ## Install requirements
	pip-sync requirements.txt dev-requirements.txt

.PHONY: lint
lint: ## Run linters
	flake8 .

requirements.txt: export CUSTOM_COMPILE_COMMAND := make requirements.txt
requirements.txt: requirements.in install-piptools
	$(call PIP_COMPILE,requirements.in)

dev-requirements.txt: export CUSTOM_COMPILE_COMMAND := make dev-requirements.txt
dev-requirements.txt: dev-requirements.in requirements.txt install-piptools
	$(call PIP_COMPILE,dev-requirements.in)

docs-requirements.txt: export CUSTOM_COMPILE_COMMAND := make docs-requirements.txt
docs-requirements.txt: docs-requirements.in install-piptools
	$(call PIP_COMPILE,docs-requirements.in)
	cp docs-requirements.txt docs/requirements.txt

.PHONY: requirements
requirements: requirements.txt dev-requirements.txt docs-requirements.txt ## Compile requirements
