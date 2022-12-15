.PHONY: flyte
flyte: FLYTE_BINARY_VERSION := databricks
flyte:
	mkdir -p images/tar/amd64
	docker pull --platform linux/amd64 \
		ghcr.io/flyteorg/flyte-binary:$(FLYTE_BINARY_VERSION)
	docker tag ghcr.io/flyteorg/flyte-binary:$(FLYTE_BINARY_VERSION) \
		flyte-binary:sandbox
	docker save --output  images/tar/amd64/flyte-binary.tar flyte-binary:sandbox
	docker rmi flyte-binary:sandbox

	mkdir -p images/tar/arm64
	docker pull --platform linux/arm64 \
		ghcr.io/flyteorg/flyte-binary:$(FLYTE_BINARY_VERSION)
	docker tag ghcr.io/flyteorg/flyte-binary:$(FLYTE_BINARY_VERSION) \
		flyte-binary:sandbox
	docker save --output  images/tar/arm64/flyte-binary.tar flyte-binary:sandbox
	docker rmi flyte-binary:sandbox

manifest:
	mkdir -p manifests
	kustomize build \
		--enable-helm \
		--load-restrictor=LoadRestrictionsNone kustomize > manifests/compiled.yaml

.PHONY: build
build: flyte manifest
	[ -n "$(shell docker buildx ls | awk '/^flyte-sandbox / {print $$1}')" ] || \
		docker buildx create --name flyte-sandbox \
		--driver docker-container --driver-opt image=moby/buildkit:master \
		--buildkitd-flags '--allow-insecure-entitlement security.insecure' \
		--platform linux/arm64,linux/amd64
	docker buildx build --builder flyte-sandbox --allow security.insecure --load \
		--tag flyte-sandbox:databricks .

.PHONY: start
start: FLYTE_SANDBOX_IMAGE := flyte-sandbox:databricks
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
