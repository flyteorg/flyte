BREW_BIN := $(shell brew --prefix)/bin
K3D_CLUSTER_NAME := flyte-sandbox
OUTPUT_MANIFEST := ../../docker/sandbox-bundled/manifests/compiled.yaml

.PHONY: start
start:
	$(BREW_BIN)/k3d cluster get $(K3D_CLUSTER_NAME) &> /dev/null || \
		$(BREW_BIN)/k3d cluster create $(K3D_CLUSTER_NAME) \
			--k3s-arg '--disable=traefik@server:0' \
			--k3s-arg '--disable=servicelb@server:0' \
			--no-lb \
			--port "30000:30000@server:0:direct" \
			--port "30001:30001@server:0:direct" \
			--port "30002:30002@server:0:direct" \
			--port "30080:30080@server:0:direct"

.PHONY: stop
stop:
	$(BREW_BIN)/k3d cluster delete $(K3D_CLUSTER_NAME)

.PHONY: render
render: start
	$(BREW_BIN)/kubectl --context=k3d-$(K3D_CLUSTER_NAME) get namespace flyte || \
	$(BREW_BIN)/kubectl --context=k3d-$(K3D_CLUSTER_NAME) create namespace flyte
	$(BREW_BIN)/helm template \
		--skip-tests \
		--namespace flyte \
		flyte-sandbox . | \
	$(BREW_BIN)/kubectl create \
		--context=k3d-$(K3D_CLUSTER_NAME) \
		--namespace=flyte \
		-f - --dry-run=client -o yaml > $(OUTPUT_MANIFEST)

.PHONY: bootstrap
bootstrap:
	$(BREW_BIN)/kubectl --context=k3d-$(K3D_CLUSTER_NAME) --namespace=flyte apply -f $(OUTPUT_MANIFEST)
