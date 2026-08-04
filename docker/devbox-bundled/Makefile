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
		--platform linux/arm64,linux/amd64

# Architectures to build the flyte-binary for. CI overrides this to the
# runner's arch only (e.g. FLYTE_ARCHES=amd64) to skip the cross build.
FLYTE_ARCHES ?= amd64 arm64

.PHONY: flyte
flyte: FLYTECONSOLE_VERSION := latest
flyte: create_builder
	$(foreach arch,$(FLYTE_ARCHES),$(call FLYTE_BINARY_BUILD,$(arch)))

.PHONY: helm-repos
helm-repos:
	helm repo add docker-registry https://twuni.github.io/docker-registry.helm
	helm repo add bitnami https://charts.bitnami.com/bitnami
	helm repo add knative-serving https://deeploy-knative-serving-charts.storage.googleapis.com/
	helm repo add rustfs https://rustfs.github.io/helm
	helm repo update

.PHONY: dep_build
dep_build: helm-repos
	cd $(SANDBOX_CHART_DIR) && helm dependency update

.PHONY: dep_update
dep_update: SANDBOX_CHART_DIR := ../../charts/flyte-devbox
dep_update: dep_build
	cd $(SANDBOX_CHART_DIR)/charts && rm -rf docker-registry flyte-binary knative-serving && for f in *.tgz; do tar xzf "$$f"; done

# Use the executor's pinned, Helm-v4-compatible kustomize instead of whatever
# kustomize the developer happens to have on PATH.
KUSTOMIZE := ../../executor/bin/kustomize

.PHONY: kustomize
kustomize:
	$(MAKE) -C ../../executor kustomize

.PHONY: manifests
manifests: dep_update kustomize
	mkdir -p manifests
	$(KUSTOMIZE) build \
		--enable-helm \
		--load-restrictor=LoadRestrictionsNone \
		kustomize/complete > manifests/complete.yaml
	@printf "\n---\n" >> manifests/complete.yaml
	@cat kustomize/net-kourier.yaml >> manifests/complete.yaml
	$(KUSTOMIZE) build \
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
	docker buildx build --builder flyte-devbox --load \
		$(BUILDX_CACHE_FLAGS) \
		$(if $(CACHEBUST),--build-arg CACHEBUST=$(CACHEBUST)) \
		--tag flyte-devbox:latest .

.PHONY: build-gpu
build-gpu: build
	docker buildx build --builder flyte-devbox --load \
		$(BUILDX_CACHE_FLAGS) \
		--file Dockerfile.gpu \
		--build-context base=docker-image://flyte-devbox:latest \
		--build-arg BASE_IMAGE=base \
		--tag flyte-devbox:gpu-latest .

override GRAFANA_NODE_PORT := 30300
override DEVBOX_NAME := flyte-devbox
override DEVBOX_CONTEXT := flyte-devbox

# Port map
# 6443 - k8s API server
# 30000 - Docker Registry
# 30001 - DB
# 30002 - RustFS
# 30080 - Flyte Proxy
# 30081 - Kourier (Knative apps)
# 30300 - Grafana (only serves something after `make monitoring`)
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
			--publish "$(GRAFANA_NODE_PORT):$(GRAFANA_NODE_PORT)" \
			$(FLYTE_DEVBOX_IMAGE); \
	fi
	@echo "Waiting for kubeconfig..."
	@# Read the kubeconfig out of the container instead of trusting the bind
	@# mount. A devbox that is already running may have been started with a
	@# different host dir at /.kube (or by other tooling), in which case a stale
	@# $(PWD)/.kube/kubeconfig sails past this wait and the next step fails with
	@# "cannot access '/.kube/kubeconfig'". Reading from the container also
	@# lands the file owned by the invoking user, so no chown is needed (the
	@# bind-mounted copy is root-owned on WSL).
	@mkdir -p $(PWD)/.kube
	@until docker exec flyte-devbox test -s /etc/rancher/k3s/k3s.yaml >/dev/null 2>&1; do sleep 1; done
	@docker exec flyte-devbox cat /etc/rancher/k3s/k3s.yaml > $(PWD)/.kube/kubeconfig
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
	kubectl patch configmap/config-autoscaler -n knative-serving --type merge \
		--patch '{"data":{"allow-zero-initial-scale":"true"}}'
	kubectl patch svc/kourier -n kourier-system --type merge \
		--patch '{"spec":{"type":"NodePort","ports":[{"name":"http2","port":80,"targetPort":8080,"nodePort":30081}]}}'

# Optional Grafana stack (collector -> Prometheus -> Grafana + the Flyte v2
# dashboard). Not bundled in the image: the three upstream images add ~1.4 GB
# unpacked, so they are pulled on demand instead. Requires a running devbox.
.PHONY: monitoring
monitoring:
	@[ -n "$$(docker ps --filter name=^$(DEVBOX_NAME)$$ --format '{{.Names}}')" ] || \
		{ echo "$(DEVBOX_NAME) is not running — run 'make devbox-run' first"; exit 1; }
	kubectl --context $(DEVBOX_CONTEXT) apply -f monitoring/monitoring.yaml
	@# Point flyte at the collector so the dashboard has something to chart.
	kubectl --context $(DEVBOX_CONTEXT) -n flyte patch cm flyte-binary-config --type merge --patch '{"data":{"200-otel.yaml":"otel:\n  type: otlpgrpc\n  otlpgrpc:\n    endpoint: http://otel-collector.monitoring.svc.cluster.local:4317\n  sampler: { parentSampler: traceid, traceIdRatio: 0.01 }\n"}}'
	kubectl --context $(DEVBOX_CONTEXT) -n flyte rollout restart deploy flyte-binary
	kubectl --context $(DEVBOX_CONTEXT) -n monitoring rollout status deploy/grafana --timeout=300s
	kubectl --context $(DEVBOX_CONTEXT) -n flyte rollout status deploy/flyte-binary --timeout=300s
	@echo
	@if docker port $(DEVBOX_NAME) | grep -q "^$(GRAFANA_NODE_PORT)/tcp"; then \
		echo "Grafana:  http://localhost:$(GRAFANA_NODE_PORT)/d/oss/flyte-v2   (anonymous admin)"; \
	else \
		echo "NOTE: this devbox predates the Grafana port being published, so"; \
		echo "      http://localhost:$(GRAFANA_NODE_PORT) is not reachable yet. Recreate it:"; \
		echo "        make devbox-stop && make devbox-run"; \
	fi
	@echo "Metrics appear ~60s after the first RPC traffic."


.PHONY: monitoring-clean
monitoring-clean:
	-kubectl --context $(DEVBOX_CONTEXT) -n flyte patch cm flyte-binary-config --type json --patch '[{"op":"remove","path":"/data/200-otel.yaml"}]'
	-kubectl --context $(DEVBOX_CONTEXT) -n flyte rollout restart deploy flyte-binary
	-kubectl --context $(DEVBOX_CONTEXT) delete -f monitoring/monitoring.yaml

.PHONY: stop
stop:
	docker stop --time 5 flyte-devbox
