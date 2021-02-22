.SILENT:

# Update PATH variable to leverage _bin directory
export PATH := .sandbox/bin:$(PATH)

# Dependencies
export DOCKER_VERSION := 20.10.3
export K3S_VERSION := v1.20.2%2Bk3s1
export KUBECTL_VERSION := v1.20.2
export FLYTE_SANDBOX_IMAGE := flyte-sandbox:latest

# Flyte sandbox configuration variables
KUBERNETES_API_PORT := 51234
FLYTE_PROXY_PORT := 30081
MINIO_PROXY_PORT := 30084
FLYTE_SANDBOX_NAME := flyte-sandbox

# Use an ephemeral kubeconfig, so as not to litter the default one
export KUBECONFIG=$(CURDIR)/.sandbox/data/config/kubeconfig

# Module of cookbook examples to register
EXAMPLES_MODULE := core

define LOG
echo "$(shell tput bold)$(shell tput setaf 2)$(1)$(shell tput sgr0)"
endef

define ERROR
echo >&2 "$(shell tput bold)$(shell tput setaf 1)$(1)! Please 'make teardown' and 'make start' again to retry.$(shell tput sgr0)"; exit 1
endef

define RUN_IN_SANDBOX
docker exec -it \
	-e MAKEFLAGS \
	-e REGISTRY=$(REGISTRY) \
	-e DOCKER_BUILDKIT=1 \
	-e VERSION=$(VERSION) \
	-w /usr/src \
	$(FLYTE_SANDBOX_NAME) \
	$(1)
endef

.PHONY: help
help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# Helper to determine if a sandbox is up and running
.PHONY: _requires-sandbox-up
_requires-sandbox-up:
ifeq ($(shell docker ps -f name=$(FLYTE_SANDBOX_NAME) --format={.ID}),)
	$(error Cluster has not been started! Use 'make start' to start a cluster)
endif

.PHONY: _prepare
_prepare:
	$(call LOG,Preparing dependencies)
	.sandbox/prepare.sh

.PHONY: start
start: _prepare  ## Start a local Flyte sandbox
	$(call LOG,Starting sandboxed Kubernetes cluster)
	docker run -d --rm --privileged --name $(FLYTE_SANDBOX_NAME) \
		-e KUBERNETES_API_PORT=$(KUBERNETES_API_PORT) \
		-e K3S_KUBECONFIG_OUTPUT=/config/kubeconfig \
		-e SANDBOX=1 \
		-e FLYTE_HOST=localhost:30081 \
		-e FLYTE_AWS_ENDPOINT=http://localhost:30084/ \
		-v $(CURDIR)/.sandbox/data/config:/config \
		-v $(CURDIR):/usr/src \
		-v /var/run \
		-p $(KUBERNETES_API_PORT):$(KUBERNETES_API_PORT) \
		-p $(FLYTE_PROXY_PORT):30081 \
		-p $(MINIO_PROXY_PORT):30084 \
		$(FLYTE_SANDBOX_IMAGE) > /dev/null
	timeout 600 sh -c "until kubectl explain deployment &> /dev/null; do sleep 1; done" || $(call ERROR,Timed out while waiting for the Kubernetes cluster to start)

	$(call LOG,Deploying Flyte)
	kubectl apply -f https://raw.githubusercontent.com/flyteorg/flyte/master/deployment/sandbox/flyte_generated.yaml
	kubectl wait --for=condition=available deployment/flyteadmin -n flyte --timeout=10m || $(call ERROR,Timed out while waiting for the Flyteadmin component to start)

	$(call LOG,Registering examples from commit: latest)
	$(call RUN_IN_SANDBOX,bash -c "REGISTRY=ghcr.io/flyteorg VERSION=latest make -C cookbook/$(EXAMPLES_MODULE) fast_register")

	kubectl wait --for=condition=available deployment/{datacatalog,flyteadmin,flyteconsole,flytepropeller} -n flyte --timeout=10m || $(call ERROR,Timed out while waiting for the Flyte deployment to start)
	$(call LOG,"Flyte deployment ready! Flyte console is now available at http://localhost:$(FLYTE_PROXY_PORT)/console")

.PHONY: teardown
teardown: _requires-sandbox-up  ## Teardown Flyte sandbox
	$(call LOG,Tearing down Flyte sandbox)
	docker rm -f -v $(FLYTE_SANDBOX_NAME) > /dev/null ||:

.PHONY: status
status: _requires-sandbox-up  ## Show status of Flyte deployment
	kubectl get pods -n flyte

.PHONY: shell
shell: _requires-sandbox-up  # Drop into a development shell
	$(call RUN_IN_SANDBOX,bash)

.PHONY: register
register: _requires-sandbox-up  ## Register Flyte cookbook workflows
	$(call LOG,Registering example workflows in cookbook/$(EXAMPLES_MODULE))
	$(call RUN_IN_SANDBOX,make -C cookbook/$(EXAMPLES_MODULE) register)

.PHONY: fast_register
fast_register: _requires-sandbox-up  ## Fast register Flyte cookbook workflows
	$(call LOG,Fast registering example workflows in cookbook/$(EXAMPLES_MODULE))
	$(call RUN_IN_SANDBOX,make -C cookbook/$(EXAMPLES_MODULE) fast_register)
