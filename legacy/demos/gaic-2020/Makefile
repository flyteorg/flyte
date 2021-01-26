IMAGE_NAME=flytesnacks
GIT_SHA=$(shell git rev-parse HEAD)
IMAGE_TAG_WITH_SHA= $(addsuffix :$(GIT_SHA),$(IMAGE_NAME))
REGISTRY=docker.io/lyft

docker_build:
	IMAGE_NAME=$(IMAGE_NAME) scripts/docker_build.sh

register-local: 
	docker run --rm --network host -e FLYTE_PLATFORM_URL='127.0.0.1:30081' $(IMAGE_TAG_WITH_SHA) pyflyte -p flytesnacks -d development -c flyte.config register workflows

docker_push:
	REGISTRY=$(REGISTRY) IMAGE_NAME=$(IMAGE_NAME) scripts/docker_build.sh

register-remote:
	docker run --rm --network host $(addsuffix /$(IMAGE_TAG_WITH_SHA),$(REGISTRY)) pyflyte -p flytesnacks -d development -c flyte-production.config register workflows
