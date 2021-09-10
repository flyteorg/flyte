.PHONY: help
help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

.PHONY: debug
debug:
	echo "IMAGE NAME ${IMAGE_NAME}"
	echo "FULL IMAGE NAME ${FULL_IMAGE_NAME}"
	echo "VERSION TAG ${VERSION}"
	echo "REGISTRY ${REGISTRY}"

TAGGED_IMAGE=${FULL_IMAGE_NAME}:${PREFIX}-${VERSION}

# This should only be used by Admins to push images to the public Dockerhub repo. Make sure you
# specify REGISTRY=ghcr.io/flyteorg or your registry before the make command otherwise this won't actually push
# Also if you want to push the docker image for sagemaker consumption then
# specify ECR_REGISTRY
.PHONY: docker_push
docker_push: docker_build
ifdef REGISTRY
	docker push ${TAGGED_IMAGE}
endif

.PHONY: fmt
fmt: # Format code with black and isort
	black .
	isort .

.PHONY: install-piptools
install-piptools:
	pip install -U pip-tools

.PHONY: setup
setup: install-piptools # Install requirements
	pip-sync dev-requirements.txt

.PHONY: lint
lint:  # Run linters
	flake8 .
