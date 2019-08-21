# WARNING: THIS FILE IS MANAGED IN THE 'BOILERPLATE' REPO AND COPIED TO OTHER REPOSITORIES.
# ONLY EDIT THIS FILE FROM WITHIN THE 'LYFT/BOILERPLATE' REPOSITORY:
# 
# TO OPT OUT OF UPDATES, SEE https://github.com/lyft/boilerplate/blob/master/Readme.rst

.PHONY: docker_build
docker_build:
	IMAGE_NAME=$$REPOSITORY ./boilerplate/lyft/docker_build/docker_build.sh

.PHONY: dockerhub_push
dockerhub_push:
	IMAGE_NAME=lyft/$$REPOSITORY REGISTRY=docker.io ./boilerplate/lyft/docker_build/docker_build.sh
