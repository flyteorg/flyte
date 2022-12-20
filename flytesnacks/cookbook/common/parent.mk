.SILENT:

# This is a Makefile you can include in directories that have child directories you want to build common flyte targets on.
SUBDIRS = $(shell ls -d */)
PWD=$(CURDIR)

.SILENT: help
.PHONY: help
help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  $(MAKE) \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: fast_serialize
fast_serialize:
	@for dir in $(SUBDIRS) ; do \
		echo "processing ${PWD}/$$dir"; \
		trimmed=$${dir%/}; \
	    if [ -f $$dir/Makefile ]; then \
	        if [ "$$dir" != 'docs/' ]; then \
        		PREFIX=$$trimmed $(MAKE) fast_serialize || exit; \
	        fi; \
		fi; \
	done

.PHONY: fast_register
fast_register: ## Registers new code changes using the last built image (assumes current HEAD refers to a built image).
	@for dir in $(SUBDIRS) ; do \
		echo "processing ${PWD}/$$dir"; \
		trimmed=$${dir%/}; \
	    if [ -f $$dir/Makefile ]; then \
	        if [ "$$dir" != 'docs/' ]; then \
        		PREFIX=$$trimmed $(MAKE) fast_register || exit; \
	        fi; \
		fi; \
	done

.PHONY: register
register: ## Builds, pushes and registers all docker images, workflows and tasks in all sub directories.
	@for dir in $(SUBDIRS) ; do \
		echo "processing ${PWD}/$$dir"; \
	    if [ -f $$dir/Makefile ]; then \
	        if [ "$$dir" != 'docs/' ]; then \
        		$(MAKE) -C $$dir register || exit; \
	        fi; \
		fi; \
	done

.PHONY: serialize
serialize: ## Builds and serializes all docker images, workflows and tasks in all sub directories.
	@for dir in $(SUBDIRS) ; do \
		echo "processing ${PWD}/$$dir"; \
		echo "$$dir"; \
	    if [ -f $$dir/Makefile ]; then \
	        if [ "$$dir" != 'docs/' ]; then \
	            $(MAKE) -C $$dir serialize || exit; \
	        fi; \
		fi; \
	done

.PHONY: docker_push
docker_push: ## Builds and pushes all docker images.
	@for dir in $(SUBDIRS) ; do \
		echo "processing ${PWD}/$$dir"; \
		test -f $$dir/Dockerfile && \
	    if [ -f $$dir/Makefile ]; then \
    		$(MAKE) -C $$dir docker_push || exit; \
		fi; \
	done

.PHONY: requirements
requirements: ## Makes all requirement files in sub directories.
	@for dir in $(SUBDIRS) ; do \
		echo "processing ${PWD}/$$dir"; \
	    if [ -f $$dir/Makefile ]; then \
	        if [ "$$dir" != 'docs/' ]; then \
        		$(MAKE) -C $$dir requirements || exit; \
	        fi; \
		fi; \
	done

.PHONY: k3d_load_image
k3d_load_image:
	@for dir in $(SUBDIRS) ; do \
		echo "processing ${PWD}/$$dir"; \
	    if [ -f $$dir/Makefile ]; then \
	        if [ "$$dir" != 'docs/' ]; then \
        		$(MAKE) -C $$dir k3d_load_image || exit; \
	        fi; \
		fi; \
	done

.PHONY: clean
clean: ## Deletes build directories (e.g. _pb_output/)
	@for dir in $(SUBDIRS) ; do \
		echo "processing ${PWD}/$$dir"; \
	    if [ -f $$dir/Makefile ]; then \
    		$(MAKE) -C $$dir clean || exit; \
		fi; \
	done
