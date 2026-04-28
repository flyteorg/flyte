# Buildx remote cache configuration.
#
# Defaults (can be overridden by setting BUILDX_CACHE_FROM / BUILDX_CACHE_TO):
#   - Inside GitHub Actions ($(GITHUB_ACTIONS) == true): use `type=gha`, so the
#     build cache is persisted in the Actions cache service across runs.
#   - Locally: use `type=local` under $(BUILDX_CACHE_DIR) so repeated local
#     builds reuse the previous build's layers.
#
# Override examples:
#   # Registry-backed cache (works anywhere):
#   make build \
#     BUILDX_CACHE_FROM="type=registry,ref=ghcr.io/<org>/flyte-devbox-cache:buildcache" \
#     BUILDX_CACHE_TO="type=registry,ref=ghcr.io/<org>/flyte-devbox-cache:buildcache,mode=max"
#
#   # Disable caching entirely:
#   make build BUILDX_CACHE_FROM= BUILDX_CACHE_TO=
#
# Note: the flyte-binary build runs once per arch (amd64 and arm64); we
# auto-scope each arch's cache separately so amd64 and arm64 don't clobber
# each other. Custom overrides apply as-is and do not get arch scoping.
BUILDX_CACHE_DIR ?= $(HOME)/.cache/flyte-devbox-buildx

ifeq ($(GITHUB_ACTIONS),true)
  _DEVBOX_CACHE_FROM := type=gha,scope=flyte-devbox
  _DEVBOX_CACHE_TO   := type=gha,scope=flyte-devbox,mode=max
  _BINARY_CACHE_FROM  = type=gha,scope=flyte-binary-$(1)
  _BINARY_CACHE_TO    = type=gha,scope=flyte-binary-$(1),mode=max
else
  _DEVBOX_CACHE_FROM := type=local,src=$(BUILDX_CACHE_DIR)/devbox
  _DEVBOX_CACHE_TO   := type=local,dest=$(BUILDX_CACHE_DIR)/devbox,mode=max
  _BINARY_CACHE_FROM  = type=local,src=$(BUILDX_CACHE_DIR)/flyte-binary-$(1)
  _BINARY_CACHE_TO    = type=local,dest=$(BUILDX_CACHE_DIR)/flyte-binary-$(1),mode=max
endif

# Track whether the user overrode the cache knobs (command line / environment)
# so we can keep per-arch scoping for the binary build when they haven't.
_CACHE_FROM_USER_SET := $(filter command environment override,$(origin BUILDX_CACHE_FROM))
_CACHE_TO_USER_SET   := $(filter command environment override,$(origin BUILDX_CACHE_TO))

BUILDX_CACHE_FROM ?= $(_DEVBOX_CACHE_FROM)
BUILDX_CACHE_TO   ?= $(_DEVBOX_CACHE_TO)

BUILDX_CACHE_FLAGS = \
	$(if $(BUILDX_CACHE_FROM),--cache-from=$(BUILDX_CACHE_FROM)) \
	$(if $(BUILDX_CACHE_TO),--cache-to=$(BUILDX_CACHE_TO))

# Per-arch cache flags for the flyte-binary build. User overrides apply as-is;
# otherwise the arch-scoped defaults keep amd64 and arm64 caches separate.
binary_cache_flags = \
	$(if $(_CACHE_FROM_USER_SET),\
		$(if $(BUILDX_CACHE_FROM),--cache-from=$(BUILDX_CACHE_FROM)),\
		--cache-from=$(call _BINARY_CACHE_FROM,$(1))) \
	$(if $(_CACHE_TO_USER_SET),\
		$(if $(BUILDX_CACHE_TO),--cache-to=$(BUILDX_CACHE_TO)),\
		--cache-to=$(call _BINARY_CACHE_TO,$(1)))

define FLYTE_BINARY_BUILD
mkdir -p images/tar/$(1)

docker buildx build \
	--build-arg FLYTECONSOLE_VERSION=$(FLYTECONSOLE_VERSION) \
	--builder flyte-devbox \
	--platform linux/$(1) \
	--tag flyte-binary-v2:sandbox \
	$(call binary_cache_flags,$(1)) \
	--output type=docker,dest=images/tar/$(1)/flyte-binary.tar \
	../..

endef

.PHONY: create_builder
create_builder:
	[ -n "$(shell docker buildx ls | awk '/^flyte-devbox / {print $$1}')" ] || \
		docker buildx create --name flyte-devbox \
		--driver docker-container --driver-opt image=moby/buildkit:master \
		--buildkitd-flags '--allow-insecure-entitlement security.insecure' \
		--platform linux/arm64,linux/amd64

.PHONY: flyte
flyte: FLYTECONSOLE_VERSION := latest
flyte: create_builder
	$(foreach arch,amd64 arm64,$(call FLYTE_BINARY_BUILD,$(arch)))

.PHONY: helm-repos
helm-repos:
	helm repo add docker-registry https://twuni.github.io/docker-registry.helm
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add knative-serving https://deeploy-knative-serving-charts.storage.googleapis.com/
	helm repo update

.PHONY: dep_build
dep_build: helm-repos
	cd $(SANDBOX_CHART_DIR) && helm dependency update

.PHONY: dep_update
dep_update: SANDBOX_CHART_DIR := ../../charts/flyte-devbox
dep_update: dep_build
	cd $(SANDBOX_CHART_DIR)/charts && rm -rf docker-registry flyte-binary knative-serving && for f in *.tgz; do tar xzf "$$f"; done

.PHONY: manifests
manifests: dep_update
	mkdir -p manifests
	kustomize build \
		--enable-helm \
		--load-restrictor=LoadRestrictionsNone \
		kustomize/complete > manifests/complete.yaml
	@printf "\n---\n" >> manifests/complete.yaml
	@cat kustomize/net-kourier.yaml >> manifests/complete.yaml
	kustomize build \
		--enable-helm \
		--load-restrictor=LoadRestrictionsNone \
		kustomize/dev > manifests/dev.yaml
	@printf "\n---\n" >> manifests/dev.yaml
	@cat kustomize/net-kourier.yaml >> manifests/dev.yaml

.PHONY: sync-crds
sync-crds:
	$(MAKE) -C ../../executor manifests

.PHONY: build
build: sync-crds flyte dep_update manifests
	docker buildx build --builder flyte-devbox --allow security.insecure --load \
		$(BUILDX_CACHE_FLAGS) \
		--tag flyte-devbox:latest .

.PHONY: build-gpu
build-gpu: build
	docker buildx build --builder flyte-devbox --allow security.insecure --load \
		$(BUILDX_CACHE_FLAGS) \
		--file Dockerfile.gpu \
		--build-context base=docker-image://flyte-devbox:latest \
		--build-arg BASE_IMAGE=base \
		--tag flyte-devbox:gpu-latest .

# Port map
# 6443 - k8s API server
# 30000 - Docker Registry
# 30001 - DB
# 30002 - RustFS
# 30080 - Flyte Proxy
# 30081 - Kourier (Knative apps)
.PHONY: start
start: FLYTE_DEVBOX_IMAGE := flyte-devbox:latest
start: FLYTE_DEV := False
start:
	[ -n "$(shell docker volume ls --filter name=^flyte-devbox$$ --format {{.Name}})" ] || \
		docker volume create flyte-devbox
	@if [ -z "$(shell docker ps --filter name=^flyte-devbox$$ --format {{.Names}})" ]; then \
		rm -f $(PWD)/.kube/kubeconfig; \
		docker run --detach --rm --privileged --name flyte-devbox \
			--add-host "host.docker.internal:host-gateway" \
			--env FLYTE_DEV=$(FLYTE_DEV) \
			--env K3S_KUBECONFIG_OUTPUT=/.kube/kubeconfig \
			--volume $(PWD)/.kube:/.kube \
				--volume flyte-devbox:/var/lib/flyte/storage \
			--publish "6443":"6443" \
			--publish "30000:30000" \
			--publish "30001:5432" \
			--publish "30002:30002" \
			--publish "30080:30080" \
			--publish "30081:30081" \
			$(FLYTE_DEVBOX_IMAGE); \
	fi
	@echo "Waiting for kubeconfig..."
	@until [ -s $(PWD)/.kube/kubeconfig ]; do sleep 1; done
	@# On WSL, the bind-mounted kubeconfig may be root-owned, which makes the host-side cp fail with permission denied.
	@docker exec flyte-devbox chown $(shell id -u):$(shell id -g) /.kube/kubeconfig
	@mkdir -p $(HOME)/.kube
	@if [ -f $(HOME)/.kube/config ]; then \
		KUBECONFIG=$(PWD)/.kube/kubeconfig:$(HOME)/.kube/config kubectl config view --flatten > /tmp/kubeconfig-merged && \
		mv /tmp/kubeconfig-merged $(HOME)/.kube/config; \
	else \
		cp $(PWD)/.kube/kubeconfig $(HOME)/.kube/config; \
	fi
	@kubectl config use-context flyte-devbox >/dev/null 2>&1 || true
	@echo "Kubeconfig merged into ~/.kube/config"
.PHONY: kubeconfig
.SILENT: kubeconfig
kubeconfig:
	sed -i -e "/server:/ s/: .*/: https:\/\/127.0.0.1:$(shell docker port flyte-devbox | grep ^6443 | awk '{print $$3}' | awk -F: '{print $$2}')/" .kube/kubeconfig
	echo "export KUBECONFIG=$(PWD)/.kube/kubeconfig"

.PHONY: setup-knative
setup-knative:
	kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-crds.yaml
	kubectl apply -f https://github.com/knative/serving/releases/download/knative-v1.21.2/serving-core.yaml
	kubectl apply -f https://github.com/knative-extensions/net-kourier/releases/download/knative-v1.21.0/kourier.yaml
	kubectl patch configmap/config-network -n knative-serving --type merge \
		--patch '{"data":{"ingress-class":"kourier.ingress.networking.knative.dev"}}'
	kubectl patch configmap/config-network -n knative-serving --type merge \
		--patch '{"data":{"domain-template":"{{.Name}}.{{.Domain}}"}}'
	kubectl patch configmap/config-domain -n knative-serving --type merge \
		--patch '{"data":{"local.flyte.app":null,"localhost":""}}'
	kubectl patch svc/kourier -n kourier-system --type merge \
		--patch '{"spec":{"type":"NodePort","ports":[{"name":"http2","port":80,"targetPort":8080,"nodePort":30081}]}}'

.PHONY: stop
stop:
	docker stop --time 5 flyte-devbox
