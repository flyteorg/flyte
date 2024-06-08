export REPOSITORY=flyte
include boilerplate/flyte/end2end/Makefile
include boilerplate/flyte/golang_test_targets/Makefile

define PIP_COMPILE
pip-compile $(1) --upgrade --verbose --resolver=backtracking --annotation-style=line
endef

GIT_VERSION := $(shell git describe --always --tags)
GIT_HASH := $(shell git rev-parse --short HEAD)
TIMESTAMP := $(shell date '+%Y-%m-%d')
PACKAGE ?=github.com/flyteorg/flytestdlib
LD_FLAGS="-s -w -X $(PACKAGE)/version.Version=$(GIT_VERSION) -X $(PACKAGE)/version.Build=$(GIT_HASH) -X $(PACKAGE)/version.BuildTime=$(TIMESTAMP)"
TMP_BUILD_DIR := .tmp_build

.PHONY: cmd/single/dist
cmd/single/dist: export FLYTECONSOLE_VERSION ?= latest
cmd/single/dist:
	script/get_flyteconsole_dist.sh

.PHONY: compile
compile: cmd/single/dist
	go build -tags console -v -o flyte -ldflags=$(LD_FLAGS) ./cmd/
	mv ./flyte ${GOPATH}/bin/ || echo "Skipped copying 'flyte' to ${GOPATH}/bin"

.PHONY: linux_compile
linux_compile: cmd/single/dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -tags console -v -o /artifacts/flyte -ldflags=$(LD_FLAGS) ./cmd/

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: helm
helm: ## Generate K8s Manifest from Helm Charts.
	bash script/generate_helm.sh
	make -C docker/sandbox-bundled manifests

.PHONY: release_automation
release_automation:
	mkdir -p release
	bash script/release.sh
	bash script/generate_config_docs.sh
	$(MAKE) -C docker/sandbox-bundled manifests

.PHONY: deploy_sandbox
deploy_sandbox:
	bash script/deploy.sh

.PHONY: install-piptools
install-piptools: ## Install pip-tools
	pip install -U pip-tools

.PHONY: install-conda-lock
install-conda-lock:
	pip install conda-lock

.PHONY: conda-lock
conda-lock: install-conda-lock
	conda-lock -f monodocs-environment.yaml --without-cuda \
		--lockfile monodocs-environment.lock.yaml \
		--platform=osx-64 --platform=osx-arm64 --platform=linux-64

.PHONY: stats
stats:
	@generate-dashboard -o deployment/stats/prometheus/flytepropeller-dashboard.json stats/flytepropeller.dashboard.py
	@generate-dashboard -o deployment/stats/prometheus/flyteadmin-dashboard.json stats/flyteadmin.dashboard.py
	@generate-dashboard -o deployment/stats/prometheus/flyteuser-dashboard.json stats/flyteuser.dashboard.py

.PHONY: prepare_artifacts
prepare_artifacts:
	bash script/prepare_artifacts.sh

.PHONE: helm_update
helm_update: ## Update helm charts' dependencies.
	helm dep update ./charts/flyte/

.PHONY: helm_install
helm_install: ## Install helm charts
	helm install flyte --debug ./charts/flyte -f ./charts/flyte/values.yaml --create-namespace --namespace=flyte

.PHONY: helm_upgrade
helm_upgrade: ## Upgrade helm charts
	helm upgrade flyte --debug ./charts/flyte -f ./charts/flyte/values.yaml --create-namespace --namespace=flyte

# Used in CI
.PHONY: docs
docs:
	make -C docs clean html SPHINXOPTS=-W

$(TMP_BUILD_DIR):
	mkdir $@

$(TMP_BUILD_DIR)/conda-lock-image: docs/Dockerfile.conda-lock | $(TMP_BUILD_DIR)
	docker buildx build --load --platform=linux/amd64 --build-arg USER_UID=$$(id -u) --build-arg USER_GID=$$(id -g) -t flyte-conda-lock:latest -f docs/Dockerfile.conda-lock .
	touch $(TMP_BUILD_DIR)/conda-lock-image

monodocs-environment.lock.yaml: monodocs-environment.yaml $(TMP_BUILD_DIR)/conda-lock-image
	docker run --platform=linux/amd64 --rm --pull never -v ./:/flyte flyte-conda-lock:latest lock --file monodocs-environment.yaml --lockfile monodocs-environment.lock.yaml

$(TMP_BUILD_DIR)/dev-docs-image: docs/Dockerfile.docs monodocs-environment.lock.yaml | $(TMP_BUILD_DIR)
	docker buildx build --load --platform=linux/amd64 --build-arg USER_UID=$$(id -u) --build-arg USER_GID=$$(id -g) -t flyte-dev-docs:latest -f docs/Dockerfile.docs .
	touch $(TMP_BUILD_DIR)/dev-docs-image

# Build docs in docker container for local development
.PHONY: dev-docs
dev-docs: $(TMP_BUILD_DIR)/dev-docs-image
	bash script/local_build_docs.sh

.PHONY: help
help: SHELL := /bin/sh
help: ## List available commands and their usage
	@awk 'BEGIN {FS = ":.*?##"; printf "\nUsage:\n  make \033[36m<target>\033[0m\n\nTargets:\n"} /^[0-9a-zA-Z_-]+:.*?##/ { printf "  \033[36m%-20s\033[0m %s\n", $$1, $$2 } ' $(MAKEFILE_LIST)

.PHONY: setup_local_dev
setup_local_dev: ## Sets up k3d cluster with Flyte dependencies for local development
	@bash script/setup_local_dev.sh

# This builds the flyte binary image for whatever architecture you're on. Don't push this image to the official
# registry because we need those to be multi-arch.
.PHONY: build_native_flyte
build_native_flyte: FLYTECONSOLE_VERSION := latest
build_native_flyte:
	docker build \
	--build-arg FLYTECONSOLE_VERSION=$(FLYTECONSOLE_VERSION) \
	--tag flyte-binary:native .

.PHONY: go-tidy
go-tidy:
	go mod tidy
	make -C datacatalog go-tidy
	make -C flyteadmin go-tidy
	make -C flyteidl go-tidy
	make -C flytepropeller go-tidy
	make -C flyteplugins go-tidy
	make -C flytestdlib go-tidy
	make -C flytecopilot go-tidy

.PHONY: lint-helm-charts
lint-helm-charts:
	# This pressuposes that you have act installed
	act pull_request -W .github/workflows/validate-helm-charts.yaml --container-architecture linux/amd64 -e charts/event.json

.PHONY: spellcheck
spellcheck:
	act pull_request --container-architecture linux/amd64 -W .github/workflows/codespell.yml

.PHONY: clean
clean: ## Remove the HTML files related to the Flyteconsole and Makefile
	rm -rf cmd/single/dist .tmp_build
