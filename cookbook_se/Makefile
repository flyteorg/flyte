include env/common/Makefile.common

.PHONY: in_container_serialize_sandbox
in_container_serialize_sandbox:
	pyflyte --config /root/sandbox.config serialize workflows -f /tmp/output

.PHONY: register_sandbox
register_sandbox: docker_build serialize_sandbox
	flyte-cli register-files -i -p ${PROJECT} -d development -v ${VERSION} -h localhost:30081 ${CURDIR}/_pb_output/*

.PHONY: serialize_sandbox
serialize_sandbox: docker_build
	echo ${CURDIR}
	mkdir ${CURDIR}/_pb_output || true
	rm ${CURDIR}/_pb_output/* || true
	docker run -v ${CURDIR}/_pb_output:/tmp/output ${FULL_IMAGE_NAME}:${VERSION} make in_container_serialize_sandbox

.PHONY: in_container_fast_serialize_sandbox
in_container_fast_serialize_sandbox:
	pyflyte --config /root/sandbox.config serialize fast workflows -f /tmp/output --source-dir /root/recipes

.PHONY: fast_register_sandbox
fast_register_sandbox: docker_build fast_serialize_sandbox
	FLYTE_AWS_ENDPOINT=http://localhost:9000/ FLYTE_AWS_ACCESS_KEY_ID=minio FLYTE_AWS_SECRET_ACCESS_KEY=miniostorage \
		flyte-cli fast-register-files -p ${PROJECT} -d development -h localhost:30081 -i \
		--additional-distribution-dir s3://my-s3-bucket/fast/ --dest-dir /root/recipes ${CURDIR}/_pb_output/*

.PHONY: fast_serialize_sandbox
fast_serialize_sandbox: docker_build
	echo ${CURDIR}
	mkdir ${CURDIR}/_pb_output || true
	rm -f ${CURDIR}/_pb_output/*.tar.gz
	docker run -v ${CURDIR}/_pb_output:/tmp/output ${FULL_IMAGE_NAME}:${VERSION} make in_container_fast_serialize_sandbox

.PHONY: enter_sandbox
enter_sandbox: docker_build
	docker run -e PROJECT=${PROJECT} -v `pwd`:/root -it ${FULL_IMAGE_NAME}:${VERSION} bash

.PHONE: run-all-examples
run-examples:
	sh ./scripts/run-all-examples.sh

PWD=$(CURDIR)
.PHONY: all_docker_push
all_docker_push: docker_push
	REGISTRY=${REGISTRY} IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh Dockerfile.sagemaker sagemaker
	REGISTRY=${REGISTRY} IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh Dockerfile.spark spark
	REGISTRY=${REGISTRY} IMAGE_NAME=${IMAGE_NAME} flytekit_build_image.sh Dockerfile.pytorch pytorch

.PHONY: all_requirements
all_requirements: requirements
	ENV_BASE_PATH="${PWD}/env/" scripts/make-all-docker.sh requirements
