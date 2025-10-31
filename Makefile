.DEFAULT_GOAL := help

# Docker CI image configuration
DOCKER_CI_IMAGE := ghcr.io/flyteorg/flyte/ci:v2

# Environment variable flags for Docker
DOCKER_ENV_FLAGS :=
ifdef GITHUB_TOKEN
	DOCKER_ENV_FLAGS += -e GITHUB_TOKEN=$(GITHUB_TOKEN)
endif
ifdef BUF_TOKEN
	DOCKER_ENV_FLAGS += -e BUF_TOKEN=$(BUF_TOKEN)
endif

DOCKER_RUN := docker run --rm -v $(CURDIR):/workspace -w /workspace -e UV_PROJECT_ENVIRONMENT=/tmp/flyte-venv $(DOCKER_ENV_FLAGS) $(DOCKER_CI_IMAGE)

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

.PHONY: buf-dep
buf-dep:
	@echo '📦  Updating buf modules (local)'
	buf dep update
	@$(MAKE) sep

.PHONY: buf-format
buf-format:
	@echo 'Running buf format (local)'
	buf format -w
	@$(MAKE) sep

.PHONY: buf-lint
buf-lint:
	@echo '🧹  Linting protocol buffer files (local)'
	buf lint --exclude-path flytestdlib/
	@$(MAKE) sep

.PHONY: buf-ts
buf-ts:
	@echo '🟦  Generating TypeScript protocol buffer files (local)'
	buf generate --clean --template buf.gen.ts.yaml --exclude-path flytestdlib/
	@cp -r flyteidl2/gen_utils/ts/* gen/ts/
	@echo '📦  Installing TypeScript dependencies'
	@cd gen/ts && npm install --silent
	@echo '✅  TypeScript generation complete'
	@$(MAKE) sep

.PHONY: buf-ts-check
buf-ts-check: buf-ts ## Generate TypeScript files and run type checking
	@echo '🔍  Type checking generated TypeScript files'
	@cd gen/ts && npx tsc --noEmit || (echo '⚠️  Type checking found issues (non-fatal)' && exit 0)
	@echo '✅  Type checking complete'
	@$(MAKE) sep

.PHONY: buf-go
buf-go:
	@echo '🟩  Generating Go protocol buffer files (local)'
	buf generate --clean --template buf.gen.go.yaml --exclude-path flytestdlib/
	@$(MAKE) sep

.PHONY: buf-rust
buf-rust:
	@echo '🦀  Generating Rust protocol buffer files (local)'
	buf generate --clean --template buf.gen.rust.yaml --exclude-path flytestdlib/
	@cp -R flyteidl2/gen_utils/rust/* gen/rust/
	@cd gen/rust && cargo update --aggressive
	@$(MAKE) sep

export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0
.PHONY: buf-python
buf-python:
	@echo '🐍  Generating Python protocol buffer files (local)'
	buf generate --clean --template buf.gen.python.yaml --exclude-path flytestdlib/
	@cp flyteidl2/gen_utils/python/* gen/python/
	@find gen/python -type d -exec touch {}/__init__.py \;
	@cd gen/python && uv lock
	@$(MAKE) sep

.PHONY: buf
buf: buf-dep buf-format buf-lint buf-rust buf-python buf-go buf-ts buf-ts-check
	@echo '🛠️  Finished generating all protocol buffer files (local)'
	@$(MAKE) sep

.PHONY: go-tidy
go-tidy:
	@echo '🧹  Running go mod tidy (local)'
	@go mod tidy $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: mocks
mocks:
	@echo "🧪  Generating go mocks (local)"
	mockery $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: gen-local
gen-local: buf mocks go-tidy ## Generate everything using local tools (requires buf, go, cargo, uv)
	@echo '⚡  Finished generating everything in the gen directory (local)'
	@$(MAKE) sep

.PHONY: build-crate
build-crate: ## Build Rust crate using local cargo
	@echo 'Cargo build the generated rust code (local)'
	cd gen/rust && cargo build
	@$(MAKE) sep

# =============================================================================
# Default Commands (use Docker - no local tools required)
# =============================================================================

.PHONY: gen
gen: ## Generate everything (uses Docker - no local tools required)
	$(DOCKER_RUN) make gen-local
	@echo '⚡  Finished generating everything in the gen directory (Docker)'
	@$(MAKE) sep

# Docker-based development targets
.PHONY: docker-pull
docker-pull: ## Pull the latest CI Docker image
	@echo '📦  Pulling latest CI Docker image'
	docker pull $(DOCKER_CI_IMAGE)
	@$(MAKE) sep

.PHONY: docker-build
docker-build: ## Build Docker CI image locally (faster iteration)
	@echo '🔨  Building Docker CI image locally (fast mode)'
	docker build -f gen.Dockerfile -t $(DOCKER_CI_IMAGE) --cache-from $(DOCKER_CI_IMAGE) .
	@echo '✅  Image built: $(DOCKER_CI_IMAGE)'
	@$(MAKE) sep

.PHONY: docker-shell
docker-shell: ## Start an interactive shell in the CI Docker container
	@echo '🐳  Starting interactive shell in CI container'
	docker run --rm -it -v $(CURDIR):/workspace -w /workspace -e UV_PROJECT_ENVIRONMENT=/tmp/flyte-venv $(DOCKER_ENV_FLAGS) $(DOCKER_CI_IMAGE) bash

# Combined workflow for fast iteration
.PHONY: docker-dev
docker-dev: docker-build gen ## Build local image and run generation (fast iteration)
	@echo '✅  Local Docker image built and generation complete!'
	@$(MAKE) sep