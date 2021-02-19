.SILENT:

include common/Makefile

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

.PHONY: all_fast_register
all_fast_register: ## Registers new code changes using the last built image (assumes current HEAD refers to a built image).
	$(MAKE) -C core fast_register
	$(MAKE) -C plugins all_fast_register
	$(MAKE) -C case_studies all_fast_register

.PHONY: all_fast_serialize
all_fast_serialize:
	$(MAKE) -C core fast_serialize
	$(MAKE) -C plugins all_fast_serialize
	$(MAKE) -C case_studies all_fast_serialize

.PHONY: all_requirements
all_requirements: dev-requirements docs-requirements ## Makes all requirements in sub directories
	$(MAKE) -C core requirements
	$(MAKE) -C plugins all_requirements
	$(MAKE) -C case_studies all_requirements

.PHONY: all_register
all_register: ## Builds, pushes and registers all docker images, workflows and tasks in all sub directories.
	$(MAKE) -j10 -C core register
	$(MAKE) -j10 -C plugins all_register
	$(MAKE) -j10 -C case_studies all_register

.PHONY: all_serialize
all_serialize: # Builds and serializes all docker images, workflows and tasks in all sub directories.
	$(MAKE) -C core serialize
	$(MAKE) -C plugins all_serialize
	$(MAKE) -C case_studies all_serialize

.PHONY: all_docker_push
all_docker_push: # Builds and pushes all docker images.
	$(MAKE) -j10 -C core docker_push
	$(MAKE) -j10 -C plugins all_docker_push
	$(MAKE) -j10 -C case_studies all_docker_push

.PHONY: all_k3d_load_image
all_k3d_load_image: 
	$(MAKE) -j10 -C core k3d_load_image
	$(MAKE) -j10 -C plugins all_k3d_load_image
	$(MAKE) -j10 -C case_studies all_k3d_load_image

.PHONY: all_clean
all_clean: # Cleans build directories (e.g. _pb_output/)
	$(MAKE) -C core clean
	$(MAKE) -C docs clean
	$(MAKE) -C plugins all_clean
	$(MAKE) -C case_studies all_clean