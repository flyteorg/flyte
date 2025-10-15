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

DOCKER_RUN := docker run --rm -v $(CURDIR):/workspace -w /workspace -e UV_PROJECT_ENVIRONMENT=/tmp/flyte-venv $(DOCKER_ENV_FLAGS) $(DOCKER_CI_IMAGE)
DOCKER_RUN_LOCAL := docker run --rm -v $(CURDIR):/workspace -w /workspace -e UV_PROJECT_ENVIRONMENT=/tmp/flyte-venv $(DOCKER_ENV_FLAGS) $(DOCKER_LOCAL_IMAGE)

SEPARATOR := \033[1;36m========================================\033[0m

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

# =============================================================================
# Local Tool Commands (require buf, go, cargo, uv installed locally)
# =============================================================================

.PHONY: buf-dep-local
buf-dep-local:
	@echo '📦  Updating buf modules (local)'
	buf dep update
	@$(MAKE) sep

.PHONY: buf-format-local
buf-format-local:
	@echo 'Running buf format (local)'
	buf format -w
	@$(MAKE) sep

.PHONY: buf-lint-local
buf-lint-local:
	@echo '🧹  Linting protocol buffer files (local)'
	buf lint --exclude-path flytestdlib/
	@$(MAKE) sep

.PHONY: buf-ts-local
buf-ts-local:
	@echo '🟦  Generating TypeScript protocol buffer files (local)'
	buf generate --clean --template buf.gen.ts.yaml --exclude-path flytestdlib/
	@cp flyteidl2/gen_utils/ts/* gen/ts/
	@$(MAKE) sep

.PHONY: buf-go-local
buf-go-local:
	@echo '🟩  Generating Go protocol buffer files (local)'
	buf generate --clean --template buf.gen.go.yaml --exclude-path flytestdlib/
	@$(MAKE) sep

.PHONY: buf-rust-local
buf-rust-local:
	@echo '🦀  Generating Rust protocol buffer files (local)'
	buf generate --clean --template buf.gen.rust.yaml --exclude-path flytestdlib/
	@cp -R flyteidl2/gen_utils/rust/* gen/rust/
	@cd gen/rust && cargo update --aggressive
	@$(MAKE) sep

export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0
.PHONY: buf-python-local
buf-python-local:
	@echo '🐍  Generating Python protocol buffer files (local)'
	buf generate --clean --template buf.gen.python.yaml --exclude-path flytestdlib/
	@cp flyteidl2/gen_utils/python/* gen/python/
	@find gen/python -type d -exec touch {}/__init__.py \;
	@cd gen/python && uv lock
	@$(MAKE) sep

.PHONY: buf-local
buf-local: buf-dep-local buf-format-local buf-lint-local buf-rust-local buf-python-local buf-go-local buf-ts-local
	@echo '🛠️  Finished generating all protocol buffer files (local)'
	@$(MAKE) sep

.PHONY: go-tidy-local
go-tidy-local:
	@echo '🧹  Running go mod tidy (local)'
	@go mod tidy $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: mocks-local
mocks-local:
	@echo "🧪  Generating go mocks (local)"
	mockery $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: gen-local
gen-local: buf-local mocks-local go-tidy-local ## Generate everything using local tools (requires buf, go, cargo, uv)
	@echo '⚡  Finished generating everything in the gen directory (local)'
	@$(MAKE) sep

.PHONY: build-crate-local
build-crate-local: ## Build Rust crate using local cargo
	@echo 'Cargo build the generated rust code (local)'
	cd gen/rust && cargo build
	@$(MAKE) sep

# =============================================================================
# Default Commands (use Docker - no local tools required)
# =============================================================================

.PHONY: buf-dep
buf-dep: ## Update buf modules (uses Docker)
	@echo '📦  Updating buf modules (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && buf dep update"
	@$(MAKE) sep

.PHONY: buf-format
buf-format: ## Format protocol buffer files (uses Docker)
	@echo 'Running buf format (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && buf format -w"
	@$(MAKE) sep

.PHONY: buf-lint
buf-lint: ## Lint protocol buffer files (uses Docker)
	@echo '🧹  Linting protocol buffer files (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && buf lint --exclude-path flytestdlib/"
	@$(MAKE) sep

.PHONY: buf-ts
buf-ts: ## Generate TypeScript protocol buffer files (uses Docker)
	@echo '🟦  Generating TypeScript protocol buffer files (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && buf generate --clean --template buf.gen.ts.yaml --exclude-path flytestdlib/ && cp flyteidl2/gen_utils/ts/* gen/ts/"
	@$(MAKE) sep

.PHONY: buf-go
buf-go: ## Generate Go protocol buffer files (uses Docker)
	@echo '🟩  Generating Go protocol buffer files (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && buf generate --clean --template buf.gen.go.yaml --exclude-path flytestdlib/"
	@$(MAKE) sep

.PHONY: buf-rust
buf-rust: ## Generate Rust protocol buffer files (uses Docker)
	@echo '🦀  Generating Rust protocol buffer files (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && buf generate --clean --template buf.gen.rust.yaml --exclude-path flytestdlib/ && cp -R flyteidl2/gen_utils/rust/* gen/rust/ && cd gen/rust && cargo update --aggressive"
	@$(MAKE) sep

.PHONY: buf-python
buf-python: ## Generate Python protocol buffer files (uses Docker)
	@echo '🐍  Generating Python protocol buffer files (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0 buf generate --clean --template buf.gen.python.yaml --exclude-path flytestdlib/ && cp flyteidl2/gen_utils/python/* gen/python/ && find gen/python -type d -exec touch {}/__init__.py \; && cd gen/python && uv lock"
	@$(MAKE) sep

.PHONY: buf
buf: buf-dep buf-format buf-lint buf-rust buf-python buf-go buf-ts ## Generate all protocol buffer files (uses Docker - recommended)
	@echo '🛠️  Finished generating all protocol buffer files (Docker)'
	@$(MAKE) sep

.PHONY: go-tidy
go-tidy: ## Run go mod tidy (uses Docker)
	@echo '🧹  Running go mod tidy (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && go mod tidy"
	@$(MAKE) sep

.PHONY: mocks
mocks: ## Generate go mocks (uses Docker)
	@echo "🧪  Generating go mocks (Docker)"
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && mockery"
	@$(MAKE) sep

.PHONY: gen
gen: buf mocks go-tidy ## Generate everything (uses Docker - no local tools required)
	@echo '⚡  Finished generating everything in the gen directory (Docker)'
	@$(MAKE) sep

.PHONY: build-crate
build-crate: ## Build Rust crate (uses Docker)
	@echo 'Cargo build the generated rust code (Docker)'
	$(DOCKER_RUN) bash -c "git config --global --add safe.directory /workspace && cd gen/rust && cargo build"
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
	docker run --rm -it -v $(CURDIR):/workspace -w /workspace -e UV_PROJECT_ENVIRONMENT=/tmp/flyte-venv $(DOCKER_ENV_FLAGS) $(DOCKER_CI_IMAGE) bash -c "git config --global --add safe.directory /workspace && bash"

.PHONY: docker-shell-local
docker-shell-local: ## Start an interactive shell in the locally built Docker container
	@echo '🐳  Starting interactive shell in local container'
	docker run --rm -it -v $(CURDIR):/workspace -w /workspace -e UV_PROJECT_ENVIRONMENT=/tmp/flyte-venv $(DOCKER_ENV_FLAGS) $(DOCKER_LOCAL_IMAGE) bash -c "git config --global --add safe.directory /workspace && bash"

.PHONY: docker-gen-local
docker-gen-local: ## Run 'make gen' inside locally built Docker container
	@echo '🐳  Running make gen in local container'
	$(DOCKER_RUN_LOCAL) bash -c "git config --global --add safe.directory /workspace && make gen-local"
	@$(MAKE) sep

# Combined workflow for fast iteration
.PHONY: docker-dev
docker-dev: docker-build docker-gen-local ## Build local image and run generation (fast iteration)
	@echo '✅  Local Docker image built and generation complete!'
	@$(MAKE) sep