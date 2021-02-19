.SILENT:

# This is a Makefile you can include in directories that have child directories you want to build common flyte targets on.
SUBDIRS = $(shell ls -d */)
PWD=$(CURDIR)

.SILENT: help
.PHONY: help
help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  $(MAKE) \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: all_fast_serialize
all_fast_serialize:
	for dir in $(SUBDIRS) ; do \
		trimmed=$${dir%/}; \
		test -f $$dir/Dockerfile && \
		PREFIX=$$trimmed $(MAKE) fast_serialize; \
	done

.PHONY: all_fast_register
all_fast_register: ## Registers new code changes using the last built image (assumes current HEAD refers to a built image).
	for dir in $(SUBDIRS) ; do \
		trimmed=$${dir%/}; \
		test -f $$dir/Dockerfile && \
		PREFIX=$$trimmed $(MAKE) fast_register; \
	done

.PHONY: all_register
all_register: ## Builds, pushes and registers all docker images, workflows and tasks in all sub directories.
	for dir in $(SUBDIRS) ; do \
		make -C  $$dir register; \
	done

.PHONY: all_serialize
all_serialize: ## Builds and serializes all docker images, workflows and tasks in all sub directories.
	echo "processing ${PWD}"
	for dir in $(SUBDIRS) ; do \
		echo "processing $$dir"; \
		make -C $$dir serialize; \
	done

.PHONY: all_docker_push
all_docker_push: ## Builds and pushes all docker images.
	echo "processing ${PWD}"
	for dir in $(SUBDIRS) ; do \
		echo "processing $$dir"; \
		make -C $$dir docker_push; \
	done

.PHONY: all_requirements
all_requirements: ## Makes all requirement files in sub directories.
	echo "processing ${PWD}"
	for dir in $(SUBDIRS) ; do \
		echo "processing $$dir"; \
		make -C $$dir requirements; \
	done

.PHONY: all_k3d_load_image
all_k3d_load_image: 
	echo "processing ${PWD}"
	for dir in $(SUBDIRS) ; do \
		echo "processing $$dir"; \
		test -f $$dir/Dockerfile && \
		test -f $$dir/Makefile && $(MAKE) -C $$dir k3d_load_image; \
	done

.PHONY: all_clean
all_clean: ## Deletes build directories (e.g. _pb_output/)
	echo "processing ${PWD}"
	for dir in $(SUBDIRS) ; do \
		echo "processing $$dir"; \
		test -f $$dir/Makefile && $(MAKE) -C $$dir clean; \
	done
