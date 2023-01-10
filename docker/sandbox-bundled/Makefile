define FLYTE_BINARY_BUILD
mkdir -p images/tar/$(1)

docker buildx build \
	--build-arg FLYTECONSOLE_VERSION=$(FLYTECONSOLE_VERSION) \
	--platform linux/$(1) \
	--tag flyte-binary:sandbox \
	--output type=docker,dest=images/tar/$(1)/flyte-binary.tar \
	../..

endef

.PHONY: flyte
flyte: FLYTECONSOLE_VERSION := latest
flyte:
	$(foreach arch,amd64 arm64,$(call FLYTE_BINARY_BUILD,$(arch)))

.PHONY: manifests
manifests:
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
build: flyte manifests
	[ -n "$(shell docker buildx ls | awk '/^flyte-sandbox / {print $$1}')" ] || \
		docker buildx create --name flyte-sandbox \
		--driver docker-container --driver-opt image=moby/buildkit:master \
		--buildkitd-flags '--allow-insecure-entitlement security.insecure' \
		--platform linux/arm64,linux/amd64
	docker buildx build --builder flyte-sandbox --allow security.insecure --load \
		--tag flyte-sandbox:latest .

.PHONY: start
start: FLYTE_SANDBOX_IMAGE := flyte-sandbox:latest
start:
	[ -n "$(shell docker ps --filter name=^flyte-sandbox$$ --format {{.Names}})" ] || \
		docker run --detach --rm --privileged --name flyte-sandbox \
			--env K3S_KUBECONFIG_OUTPUT=/.kube/kubeconfig \
			--volume $(PWD)/.kube:/.kube \
			--publish "6443" \
			--publish "30000:30000" \
			--publish "30001:30001" \
			--publish "30002:30002" \
			--publish "30080:30080" \
			$(FLYTE_SANDBOX_IMAGE)

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

.PHONY: test
test: export FLYTECTL_CONFIG := test/config.yaml
test: .venv
	. .venv/bin/activate && pyflyte run --remote test/wf.py wf --n 500 --mean 42 --sigma 2
