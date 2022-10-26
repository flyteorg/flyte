.SILENT: help
.PHONY: help
help:
	echo Available recipes:
	cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' | awk 'BEGIN { FS = ":.*?## " } { cnt++; a[cnt] = $$1; b[cnt] = $$2; if (length($$1) > max) max = length($$1) } END { for (i = 1; i <= cnt; i++) printf "  $(shell tput setaf 6)%-*s$(shell tput setaf 0) %s\n", max, a[i], b[i] }'
	tput sgr0

export DOCKER_BUILDKIT := 1

.PHONY: images
images:
	mkdir -p $(dir $@)
	docker pull ghcr.io/flyteorg/flyte-binary:$(VERSION) --platform linux/arm64 $(1)
	docker tag ghcr.io/flyteorg/flyte-binary:$(VERSION) ghcr.io/flyteorg/flyte-binary:latest
	docker save --output  $(dir $@)arm64/flyte.tar ghcr.io/flyteorg/flyte-binary:latest
	docker rmi ghcr.io/flyteorg/flyte-binary
	docker pull ghcr.io/flyteorg/flyte-binary:$(VERSION) --platform linux/amd64 $(1)
	docker tag ghcr.io/flyteorg/flyte-binary:$(VERSION) ghcr.io/flyteorg/flyte-binary:latest
	docker save --output  $(dir $@)amd64/flyte.tar ghcr.io/flyteorg/flyte-binary:latest

.PHONY: images-ci
images-ci:
	mkdir -p $(dir $@)amd64
	mkdir -p $(dir $@)arm64
	cp /tmp/flyte-amd64.tar $(dir $@)amd64/flyte.tar
	cp /tmp/flyte-arm64.tar $(dir $@)arm64/flyte.tar

.PHONY: clean
clean: ## Cleanup state
	rm -rf data

.PHONY: build
build:  images
	docker build -t flyte-sandbox -f images/dockerfiles/k3s/Dockerfile .

data/flyte.yaml:
	mkdir -p $(shell dirname $@)
	cp bootstrap/flyte/flyte.yaml $@

.PHONY: start
start:  ## Start sandbox
	[ -n "$(shell docker ps --filter name=^flyte-sandbox$$ --format {{.Names}})" ] || \
		docker run --detach --rm --privileged --name flyte-sandbox --volume $(PWD)/data:/srv/flyte --env K3S_KUBECONFIG_OUTPUT=/srv/flyte/kubeconfig --publish "6443" --publish "30000:30000" --publish "30001:30001" --publish "30002:30002" --publish "30080:30080" docker.io/library/flyte-sandbox

.PHONY: kubeconfig
.SILENT: kubeconfig
kubeconfig: ## Set up kubeconfig
	sed -i -e "/server:/ s/: .*/: https:\/\/127.0.0.1:$(shell docker port flyte-sandbox | grep ^6443 | awk '{print $$3}' | awk -F: '{print $$2}')/" data/kubeconfig
	echo "export KUBECONFIG=$(PWD)/data/kubeconfig"

.PHONY: stop
stop: ## Stop sandbox
	docker stop flyte-sandbox

.PHONY: console
console: ## Open Flyte console in browser
	open http://localhost:30080/console

.venv:
	python -m venv .venv
	. .venv/bin/activate && pip install flytekit

.PHONY: example
example: export FLYTECTL_CONFIG := example/config.yaml
example: .venv ## Run example script
	. .venv/bin/activate && pyflyte run --remote example/example.py wf --n 500 --mean 42 --sigma 2
