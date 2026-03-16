define FLYTE_BINARY_BUILD
mkdir -p images/tar/$(1)

docker buildx build \
	--build-arg FLYTECONSOLE_VERSION=$(FLYTECONSOLE_VERSION) \
	--builder flyte-sandbox \
	--platform linux/$(1) \
	--tag flyte-binary-v2:sandbox \
	--output type=docker,dest=images/tar/$(1)/flyte-binary.tar \
	../..

endef

.PHONY: create_builder
create_builder:
	[ -n "$(shell docker buildx ls | awk '/^flyte-sandbox / {print $$1}')" ] || \
		docker buildx create --name flyte-sandbox \
		--driver docker-container --driver-opt image=moby/buildkit:master \
		--buildkitd-flags '--allow-insecure-entitlement security.insecure' \
		--platform linux/arm64,linux/amd64

.PHONY: flyte
flyte: FLYTECONSOLE_VERSION := latest
flyte: create_builder
	$(foreach arch,amd64 arm64,$(call FLYTE_BINARY_BUILD,$(arch)))

.PHONY: helm-repos
helm-repos:
	helm repo add docker-registry https://twuni.github.io/docker-registry.helm || true
	helm repo add kubernetes-dashboard https://kubernetes-retired.github.io/dashboard/ || true
	helm repo add bitnami https://charts.bitnami.com/bitnami || true
	helm repo update

.PHONY: dep_build
dep_build: helm-repos
	cd $(SANDBOX_CHART_DIR) && helm dependency build

.PHONY: dep_update
dep_update: SANDBOX_CHART_DIR := ../../charts/flyte-sandbox
dep_update: dep_build
	cd $(SANDBOX_CHART_DIR)/charts && for f in *.tgz; do tar xzf "$$f"; done

.PHONY: manifests
manifests: dep_update
	mkdir -p manifests
	kustomize build \
		--enable-helm \
		--load-restrictor=LoadRestrictionsNone \
		kustomize/complete > manifests/complete.yaml
	kustomize build \
		--enable-helm \
		--load-restrictor=LoadRestrictionsNone \
		kustomize/dev > manifests/dev.yaml

.PHONY: build
build: flyte dep_update manifests
	docker buildx build --builder flyte-sandbox --allow security.insecure --load \
		--tag flyte-sandbox-v2:latest .

# Port map
# 6443 - k8s API server
# 30000 - Docker Registry
# 30001 - DB
# 30002 - Minio
# 30003 - Buildkit
# 30080 - Flyte Proxy
.PHONY: start
start: FLYTE_SANDBOX_IMAGE := flyte-sandbox-v2:latest
start: FLYTE_DEV := False
start:
	[ -n "$(shell docker volume ls --filter name=^flyte-sandbox$$ --format {{.Name}})" ] || \
		docker volume create flyte-sandbox
	@if [ -z "$(shell docker ps --filter name=^flyte-sandbox$$ --format {{.Names}})" ]; then \
		rm -f $(PWD)/.kube/kubeconfig; \
		docker run --detach --rm --privileged --name flyte-sandbox \
			--add-host "host.docker.internal:host-gateway" \
			--env FLYTE_DEV=$(FLYTE_DEV) \
			--env K3S_KUBECONFIG_OUTPUT=/.kube/kubeconfig \
			--volume $(PWD)/.kube:/.kube \
			--volume $(HOME)/.flyte/sandboxv2:/var/lib/flyte/config \
			--volume flyte-sandbox:/var/lib/flyte/storage \
			--publish "6443":"6443" \
			--publish "30000:30000" \
			--publish "30001:30001" \
			--publish "30002:30002" \
			--publish "30003:30003" \
			--publish "30080:30080" \
			$(FLYTE_SANDBOX_IMAGE); \
	fi
	@echo "Waiting for kubeconfig..."
	@until [ -s $(PWD)/.kube/kubeconfig ]; do sleep 1; done
	@# On WSL, the bind-mounted kubeconfig may be root-owned, which makes the host-side cp fail with permission denied.
	@docker exec flyte-sandbox chown $(shell id -u):$(shell id -g) /.kube/kubeconfig
	@mkdir -p $(HOME)/.flyte/sandboxv2
	@cp $(PWD)/.kube/kubeconfig $(HOME)/.flyte/sandboxv2/kubeconfig
	@echo "Kubeconfig copied to ~/.flyte/sandboxv2/kubeconfig"
.PHONY: kubeconfig
.SILENT: kubeconfig
kubeconfig:
	sed -i -e "/server:/ s/: .*/: https:\/\/127.0.0.1:$(shell docker port flyte-sandbox | grep ^6443 | awk '{print $$3}' | awk -F: '{print $$2}')/" .kube/kubeconfig
	echo "export KUBECONFIG=$(PWD)/.kube/kubeconfig"

.PHONY: stop
stop:
	docker stop flyte-sandbox

.PHONY: console
console:
	open http://localhost:30080/console

.venv:
	python -m venv .venv
	. .venv/bin/activate && pip install flytekit
