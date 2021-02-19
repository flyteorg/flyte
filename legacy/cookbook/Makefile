# This is used by the image building script referenced below. Normally it just takes the directory name but in this
# case we want it to be called something else.
IMAGE_NAME=flytecookbook
VERSION=$(shell ./version.sh)

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

.PHONY: docker_build
docker_build:
	NOPUSH=1 IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh .

# This should only be used by Admins to push images to the public Dockerhub repo. Make sure you
# specify REGISTRY=docker.io/lyft before the make command otherwise this won't actually push
.PHONY: docker_push
docker_push:
	IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh .

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

.PHONY: run_tests
run_tests: 
	export PYTHONPATH=.:$(PYTHONPATH)
	pytest tests/
