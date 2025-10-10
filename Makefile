.DEFAULT_GOAL := help

# Docker CI image configuration
DOCKER_CI_IMAGE := ghcr.io/flyteorg/flyte/ci:v2
DOCKER_LOCAL_IMAGE := flyte-ci:local

# Environment variable flags for Docker
DOCKER_ENV_FLAGS :=
ifdef GITHUB_TOKEN
	DOCKER_ENV_FLAGS += -e GITHUB_TOKEN=$(GITHUB_TOKEN)
endif
ifdef BUF_TOKEN
	DOCKER_ENV_FLAGS += -e BUF_TOKEN=$(BUF_TOKEN)
endif

DOCKER_RUN := docker run --rm -v $(CURDIR):/workspace -w /workspace $(DOCKER_ENV_FLAGS) $(DOCKER_CI_IMAGE)
DOCKER_RUN_LOCAL := docker run --rm -v $(CURDIR):/workspace -w /workspace $(DOCKER_ENV_FLAGS) $(DOCKER_LOCAL_IMAGE)

SEPARATOR := \033[1;36m========================================\033[0m
ifeq ($(VERBOSE),1)
	OUT_REDIRECT =
else
	OUT_REDIRECT = > /dev/null
endif

.PHONY: help
help: ## Show this help message
	@echo '🆘  Showing help message'
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: sep
sep:
	@echo "$(SEPARATOR)"

.PHONY: buf-dep
buf-dep: ## Update buf modules
	@echo '📦  Updating buf modules'
	buf dep update $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: buf-format
buf-format:
	@echo 'Running buf format'
	buf format -w
	@$(MAKE) sep

.PHONY: buf
buf: buf-dep buf-format buf-lint buf-rust buf-python buf-go buf-ts ## Generate all protocol buffer files for all languages
	@echo '🛠️  Finished generating all protocol buffer files for all languages'
	@$(MAKE) sep

.PHONY: buf-lint
buf-lint: ## Lint protocol buffer files
	@echo '🧹  Linting protocol buffer files'
	buf lint --exclude-path flytestdlib/ $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: buf-ts
buf-ts: ## Generate TypeScript protocol buffer files
	@echo '🟦  Generating TypeScript protocol buffer files'
	buf generate --clean --template buf.gen.ts.yaml --exclude-path flytestdlib/ $(OUT_REDIRECT)
	@cp flyteidl2/gen_utils/ts/* gen/ts/
	@$(MAKE) sep

.PHONY: buf-go
buf-go: ## Generate Go protocol buffer files
	@echo '🟩  Generating Go protocol buffer files'
	buf generate --clean --template buf.gen.go.yaml --exclude-path flytestdlib/ $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: buf-rust
buf-rust: ## Generate Rust protocol buffer files
	@echo '🦀  Generating Rust protocol buffer files'
	buf generate --clean --template buf.gen.rust.yaml --exclude-path flytestdlib/ $(OUT_REDIRECT)
	@cp -R flyteidl2/gen_utils/rust/* gen/rust/
	@cd gen/rust && cargo update --aggressive
	@$(MAKE) sep

export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0
.PHONY: buf-python
buf-python: ## Generate Python protocol buffer files
	@echo '🐍  Generating Python protocol buffer files'
	buf generate --clean --template buf.gen.python.yaml --exclude-path flytestdlib/ $(OUT_REDIRECT)
	@cp flyteidl2/gen_utils/python/* gen/python/
	@find gen/python -type d -exec touch {}/__init__.py \;
	@cd gen/python && uv lock
	@$(MAKE) sep

.PHONY: go_tidy
go_tidy: ## Run go mod tidy
	@echo '🧹  Running go mod tidy'
	@go mod tidy $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: go-tidy
go-tidy: go_tidy ## Run go mod tidy

.PHONY: download_tooling
download_tooling: ## Download necessary tooling (mockery, protoc-gen-go, etc.)
	@echo '⬇️  Downloading necessary tooling'
	go install github.com/vektra/mockery/v2@v2.53.5
	@$(MAKE) sep

.PHONY: mocks
mocks: ## Generate go mocks
	@echo "🧪  Generating go mocks"
	mockery $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: gen
gen: buf mocks go_tidy ## Generates everything in the 'gen' directory
	@echo '⚡  Finished generating everything in the gen directory'
	@$(MAKE) sep

build-crate: ## Builds the rust crate
	@echo 'Cargo build the generated rust code'
	cd gen/rust && cargo build
	@$(MAKE) sep

# Docker-based development targets
.PHONY: docker-pull
docker-pull: ## Pull the latest CI Docker image
	@echo '📦  Pulling latest CI Docker image'
	docker pull $(DOCKER_CI_IMAGE)
	@$(MAKE) sep

.PHONY: docker-build
docker-build: ## Build Docker CI image locally (faster iteration)
	@echo '🔨  Building Docker CI image locally'
	docker build -f ci.Dockerfile -t $(DOCKER_LOCAL_IMAGE) .
	@echo '✅  Image built: $(DOCKER_LOCAL_IMAGE)'
	@$(MAKE) sep

.PHONY: docker-build-fast
docker-build-fast: ## Build Docker CI image locally with cache (no pull)
	@echo '🔨  Building Docker CI image locally (fast mode)'
	docker build -f ci.Dockerfile -t $(DOCKER_LOCAL_IMAGE) --cache-from $(DOCKER_LOCAL_IMAGE) .
	@echo '✅  Image built: $(DOCKER_LOCAL_IMAGE)'
	@$(MAKE) sep

.PHONY: docker-shell
docker-shell: ## Start an interactive shell in the CI Docker container
	@echo '🐳  Starting interactive shell in CI container'
	docker run --rm -it -v $(CURDIR):/workspace -w /workspace $(DOCKER_ENV_FLAGS) $(DOCKER_CI_IMAGE) bash

.PHONY: docker-shell-local
docker-shell-local: ## Start an interactive shell in the locally built Docker container
	@echo '🐳  Starting interactive shell in local container'
	docker run --rm -it -v $(CURDIR):/workspace -w /workspace $(DOCKER_ENV_FLAGS) $(DOCKER_LOCAL_IMAGE) bash

.PHONY: docker-gen
docker-gen: ## Run 'make gen' inside Docker container
	@echo '🐳  Running make gen in CI container'
	$(DOCKER_RUN) make gen
	@$(MAKE) sep

.PHONY: docker-gen-local
docker-gen-local: ## Run 'make gen' inside locally built Docker container
	@echo '🐳  Running make gen in local container'
	$(DOCKER_RUN_LOCAL) make gen
	@$(MAKE) sep

.PHONY: docker-build-crate
docker-build-crate: ## Build Rust crate inside Docker container
	@echo '🐳  Building Rust crate in CI container'
	$(DOCKER_RUN) make build-crate
	@$(MAKE) sep

.PHONY: docker-build-crate-local
docker-build-crate-local: ## Build Rust crate inside locally built Docker container
	@echo '🐳  Building Rust crate in local container'
	$(DOCKER_RUN_LOCAL) make build-crate
	@$(MAKE) sep

# Combined workflow for fast iteration
.PHONY: docker-dev
docker-dev: docker-build docker-gen-local ## Build local image and run generation (fast iteration)
	@echo '✅  Local Docker image built and generation complete!'
	@$(MAKE) sep