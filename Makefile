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

.PHONY: cmd/single/dist
cmd/single/dist: export FLYTECONSOLE_VERSION ?= latest
cmd/single/dist:
	script/get_flyteconsole_dist.sh

.PHONY: compile
compile: cmd/single/dist
	go build -tags console -v -o flyte -ldflags=$(LD_FLAGS) ./cmd/
	mv ./flyte ${GOPATH}/bin || echo "Skipped copying 'flyte' to ${GOPATH}/bin"

.PHONY: linux_compile
linux_compile: cmd/single/dist
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -tags console -v -o /artifacts/flyte -ldflags=$(LD_FLAGS) ./cmd/

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: kustomize
kustomize:
	KUSTOMIZE_VERSION=3.9.2 bash script/generate_kustomize.sh

.PHONY: helm
helm: ## Generate K8s Manifest from Helm Charts.
	bash script/generate_helm.sh

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

.PHONY: doc-requirements.txt
doc-requirements.txt: doc-requirements.in install-piptools
	$(call PIP_COMPILE,doc-requirements.in)

.PHONY: install-conda-lock
install-conda-lock:
	pip install conda-lock

.PHONY: conda-lock
conda-lock: install-conda-lock
	conda-lock -f monodocs-environment.yaml --without-cuda --lockfile monodocs-environment.lock.yaml

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

.PHONY: docs
docs:
	make -C docs clean html SPHINXOPTS=-W

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
	$(MAKE) -C datacatalog go-tidy
	$(MAKE) -C flyteadmin go-tidy
	$(MAKE) -C flyteidl go-tidy
	$(MAKE) -C flytepropeller go-tidy
	$(MAKE) -C flyteplugins go-tidy
	$(MAKE) -C flytestdlib go-tidy
	$(MAKE) -C flytecopilot go-tidy

.PHONY: lint
lint-all:
	$(MAKE) -C datacatalog lint
	$(MAKE) -C flyteadmin lint
	$(MAKE) -C flyteidl lint
	$(MAKE) -C flytepropeller lint
	$(MAKE) -C flyteplugins lint
	$(MAKE) -C flytestdlib lint
	$(MAKE) -C flytecopilot lint


.PHONY: download_tooling
download_tooling: #download dependencies (including test deps) for the package
	@boilerplate/flyte/golang_test_targets/download_tooling.sh

.PHONY: lint
lint: download_tooling #lints the package for common code smells
	GL_DEBUG=linters_output,env find . -type f -name go.mod -maxdepth 3 -execdir golangci-lint run --deadline=5m --exclude deprecated -v \;
