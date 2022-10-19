.SILENT: help
.PHONY: help
help:
	echo Available recipes:
	cat $(MAKEFILE_LIST) | grep -E '^[a-zA-Z0-9_-]+:.*?## .*$$' | awk 'BEGIN { FS = ":.*?## " } { cnt++; a[cnt] = $$1; b[cnt] = $$2; if (length($$1) > max) max = length($$1) } END { for (i = 1; i <= cnt; i++) printf "  $(shell tput setaf 6)%-*s$(shell tput setaf 0) %s\n", max, a[i], b[i] }'
	tput sgr0

export DOCKER_BUILDKIT := 1

IMAGES := \
	kubernetesui/dashboard:v2.6.1 \
	bitnami/minio:2022-debian-11 \
	postgres:14-alpine \
	envoyproxy/envoy:v1.23-latest \
	registry:2

define IMAGE_RULE
images/tar/$(subst /,-,$(word 1,$(subst :, ,$(1)))):
	mkdir -p $$(dir $$@)/arm64
	docker pull --platform linux/arm64 $(1)
	docker tag $(1) $(word 1,$(subst :, ,$(1))):latest
	docker save --output $$(dir $$@)arm64/$(subst /,-,$(word 1,$(subst :, ,$(1)))).tar $(word 1,$(subst :, ,$(1))):latest

	mkdir -p $$(dir $$@)/amd64
	docker pull --platform linux/amd64 $(1)
	docker tag $(1) $(word 1,$(subst :, ,$(1))):latest
	docker save --output $$(dir $$@)amd64/$(subst /,-,$(word 1,$(subst :, ,$(1)))).tar $(word 1,$(subst :, ,$(1))):latest
endef

$(foreach name,$(IMAGES),$(eval $(call IMAGE_RULE,$(name))))

images/tar/flyte:
	mkdir -p $(dir $@)amd64
	mkdir -p $(dir $@)arm64
	docker build  -t flyte ../../ -f ../../Dockerfile  --platform linux/amd64
	docker save --output $(dir $@)amd64/flyte.tar flyte:latest
	docker build  -t flyte ../../ -f ../../Dockerfile  --platform linux/arm64
	docker save --output $(dir $@)arm64/flyte.tar flyte:latest

.PHONY: images
images: $(foreach name,flyte $(IMAGES),images/tar/$(subst /,-,$(word 1,$(subst :, ,$(name)))))

.PHONY: bootstrap
bootstrap:
	kubectl kustomize bootstrap > bootstrap/output.yaml

.PHONY: clean
clean: ## Cleanup state
	rm -rf data

.PHONY: build
build: bootstrap images
	docker build -t flyte-sandbox -f images/dockerfiles/k3s/Dockerfile .

data/flyte.yaml:
	mkdir -p $(shell dirname $@)
	cp bootstrap/flyte/flyte.yaml $@

.PHONY: start
start: data/flyte.yaml ## Start sandbox
	[ -n "$(shell docker ps --filter name=^flyte-sandbox$$ --format {{.Names}})" ] || \
		docker run --detach --rm --privileged --name flyte-sandbox \
			--volume $(PWD)/data:/srv/flyte \
			--env K3S_KUBECONFIG_OUTPUT=/srv/flyte/kubeconfig \
			--publish "6443" \
			--publish "30000:30000" \
			--publish "30001:30001" \
			--publish "30002:30002" \
			--publish "30080:30080" \
			flyte-sandbox

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
