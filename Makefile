export REPOSITORY=flytesnacks

.SILENT:

# Flyte sandbox configuration variables
KUBERNETES_API_PORT := 30086
FLYTE_PROXY_PORT := 30081
K8S_DASHBOARD_PROXY_PORT := 30082
MINIO_PROXY_PORT := 30084
FLYTE_SANDBOX_NAME := flyte-sandbox
KUBE_CONFIG := ~/kubeconfig

# Module of cookbook examples to register
EXAMPLES_MODULE := core

define LOG
echo "$(shell tput bold)$(shell tput setaf 2)$(1)$(shell tput sgr0)"
endef

define RUN_IN_SANDBOX
docker exec -it \
	-e DOCKER_BUILDKIT=1 \
	-e MAKEFLAGS \
	-e REGISTRY \
	-e VERSION \
	-w /usr/src \
	$(FLYTE_SANDBOX_NAME) \
	$(1)
endef

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: help
help: ## show help message
	@awk 'BEGIN {FS = ":.*##"; printf "\nUsage:\n  make \033[36m\033[0m\n"} /^[$$()% a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-15s\033[0m %s\n", $$1, $$2 } /^##@/ { printf "\n\033[1m%s\033[0m\n", substr($$0, 5) } ' $(MAKEFILE_LIST)

# Helper to determine if a sandbox is up and running
.PHONY: _requires-sandbox-up
_requires-sandbox-up:
ifeq ($(shell docker ps -f name=$(FLYTE_SANDBOX_NAME) --format={.ID}),)
	$(error Cluster has not been started! Use 'make start' to start a cluster)
endif

.PHONY: setup
setup:
	$(call LOG,Starting Flyte sandbox)
	docker run -d --rm --privileged --name $(FLYTE_SANDBOX_NAME) \
		-e SANDBOX=1 \
		-e KUBERNETES_API_PORT=$(KUBERNETES_API_PORT) \
		-e FLYTE_HOST=localhost:30081 \
		-e FLYTE_AWS_ENDPOINT=http://localhost:30084/ \
		-v $(CURDIR):/usr/src \
		-v $(KUBE_CONFIG):/etc/rancher/ \
		-p $(KUBERNETES_API_PORT):$(KUBERNETES_API_PORT) \
		-p $(FLYTE_PROXY_PORT):30081 \
		-p $(K8S_DASHBOARD_PROXY_PORT):30082 \
		-p $(MINIO_PROXY_PORT):30084 \
		ghcr.io/flyteorg/flyte-sandbox:dind > /dev/null

.PHONY: wait
wait:
	$(call RUN_IN_SANDBOX, wait-for-flyte.sh)

## Start a local Flyte sandbox
.PHONY: start
start: setup wait
	$(call LOG,Registering examples from commit: latest)
	REGISTRY=ghcr.io/flyteorg VERSION=latest $(call RUN_IN_SANDBOX,make -C cookbook/$(EXAMPLES_MODULE) fast_register)

	echo "Flyte is ready! Flyte UI is available at http://localhost:$(FLYTE_PROXY_PORT)/console."

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

.PHONY: setup-kubectl
kubectl-config: 
	# In shell/bash, run: `eval $(make kubectl-config)`
	# Makefiles run recipes in sub-processes. A sub-process cannot modify the parent process's environment.
	# The best I (@EngHabu) can think of at the moment is to output this for the user to eval in the
	# parent process.
	echo "export KUBECONFIG=$(KUBECONFIG):~/.kube/config:$(KUBE_CONFIG)/k3s/k3s.yaml"
