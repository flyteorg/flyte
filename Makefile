.DEFAULT_GOAL := help

CLUSTER_NAME ?= flytev2

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

# Include common Go targets
include go.Makefile

# =============================================================================
# Go Services Build
# =============================================================================

.PHONY: build
build: verify ## Build all Go service binaries
	$(MAKE) -C manager build
	$(MAKE) -C runs build
	$(MAKE) -C executor build

# =============================================================================
# Devbox Commands
# =============================================================================

.PHONY: devbox-build
devbox-build: ## Build the flyte devbox image (docker/devbox-bundled)
	$(MAKE) -C docker/devbox-bundled build

# Run in dev mode with extra arg FLYTE_DEV=True
.PHONY: devbox-run
devbox-run: ## Start the flyte devbox (Knative is pre-baked into the image)
	$(MAKE) -C docker/devbox-bundled start FLYTE_DEV=$(FLYTE_DEV)

.PHONY: devbox-stop
devbox-stop: ## Stop the flyte devbox
	$(MAKE) -C docker/devbox-bundled stop

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

# Helper to time a step: $(call timed,step_name,command)
define timed
	@start=$$(date +%s); \
	$(2); \
	elapsed=$$((  $$(date +%s) - $$start )); \
	echo "⏱  $(1) completed in $${elapsed}s"
endef

.PHONY: buf-dep
buf-dep:
	@echo '📦  Updating buf modules (local)'
	$(call timed,buf-dep,buf dep update)
	@$(MAKE) sep

.PHONY: buf-format
buf-format:
	@echo 'Running buf format (local)'
	$(call timed,buf-format,buf format -w)
	@$(MAKE) sep

.PHONY: buf-lint
buf-lint:
	@echo '🧹  Linting protocol buffer files (local)'
	$(call timed,buf-lint,buf lint --exclude-path flytestdlib/)
	@$(MAKE) sep

.PHONY: buf-ts
buf-ts:
	@echo '🟦  Generating TypeScript protocol buffer files (local)'
	$(call timed,buf-ts,buf generate --clean --template buf.gen.ts.yaml --exclude-path flytestdlib/ && \
		cp -r flyteidl2/gen_utils/ts/* gen/ts/ && \
		echo '📦  Installing TypeScript dependencies' && \
		cd gen/ts && npm install --silent && \
		echo '✅  TypeScript generation complete')
	@$(MAKE) sep

.PHONY: buf-ts-check
buf-ts-check: buf-ts
	@echo '🔍  Type checking generated TypeScript files'
	$(call timed,buf-ts-check,cd gen/ts && npx tsc --noEmit || (echo '⚠️  Type checking found issues (non-fatal)' && exit 0))
	@echo '✅  Type checking complete'
	@$(MAKE) sep

.PHONY: buf-go
buf-go:
	@echo '🟩  Generating Go protocol buffer files (local)'
	$(call timed,buf-go,buf generate --clean --template buf.gen.go.yaml --exclude-path flytestdlib/)
	@$(MAKE) sep

.PHONY: buf-rust
buf-rust:
	@echo '🦀  Generating Rust protocol buffer files (local)'
	$(call timed,buf-rust,buf generate --clean --template buf.gen.rust.yaml --exclude-path flytestdlib/ && \
		cp -R flyteidl2/gen_utils/rust/* gen/rust/ && \
		cd gen/rust && cargo update)
	@$(MAKE) sep

export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0
.PHONY: buf-python
buf-python:
	@echo '🐍  Generating Python protocol buffer files (local)'
	$(call timed,buf-python,buf generate --clean --template buf.gen.python.yaml --exclude-path flytestdlib/ && \
		cp flyteidl2/gen_utils/python/* gen/python/ && \
		find gen/python -type d -exec touch {}/__init__.py \; && \
		cd gen/python && uv lock)
	@$(MAKE) sep

.PHONY: buf
buf: buf-dep buf-format buf-lint buf-rust buf-python buf-go buf-ts buf-ts-check
	@echo '🛠️  Finished generating all protocol buffer files (local)'
	@$(MAKE) sep

.PHONY: go-tidy
go-tidy:
	@echo '🧹  Running go mod tidy (local)'
	$(call timed,go-tidy,go mod tidy $(OUT_REDIRECT))
	@$(MAKE) sep

.PHONY: mocks
mocks:
	@echo "🧪  Generating go mocks (local)"
	$(call timed,mocks,mockery $(OUT_REDIRECT))
	@$(MAKE) sep

.PHONY: gen-local
gen-local: buf mocks go-tidy ## Generate everything using local tools (requires buf, go, cargo, uv)
	@echo '⚡  Finished generating everything in the gen directory (local)'
	@$(MAKE) sep

.PHONY: check-crate
check-crate: ## Verify Rust crate compiles using local cargo (faster, no artifacts)
	@echo 'Cargo check the generated rust code (local)'
	cd gen/rust && cargo check

.PHONY: build-crate
build-crate: ## Build Rust crate using local cargo
	@echo 'Cargo build the generated rust code (local)'
	cd gen/rust && cargo build
	@$(MAKE) sep

# =============================================================================
# Package Dry-Run Commands (validate packages before publishing)
# =============================================================================

.PHONY: dry-run-npm
dry-run-npm: ## Dry-run npm package (shows what will be published)
	@echo '📦  NPM Package Dry Run'
	@echo '─────────────────────────────────────────'
	@echo '📄  Package files that will be included:'
	@cd gen/ts && npm pack --dry-run 2>&1 | grep -v "npm notice" || true
	@echo ''
	@echo '📋  Package contents (from package.json "files" field):'
	@cd gen/ts && cat package.json | grep -A 10 '"files"'
	@echo ''
	@echo '✅  Validation: Running npm pack to create tarball...'
	@cd gen/ts && npm pack
	@echo ''
	@echo '📦  Contents of generated tarball:'
	@cd gen/ts && tar -tzf flyteorg-flyteidl2-*.tgz | head -50
	@echo ''
	@echo '🧹  Cleaning up tarball...'
	@cd gen/ts && rm -f flyteorg-flyteidl2-*.tgz
	@echo '✅  NPM dry run complete!'
	@$(MAKE) sep

.PHONY: dry-run-python
dry-run-python: ## Dry-run Python package (shows what will be published)
	@echo '🐍  Python Package Dry Run'
	@echo '─────────────────────────────────────────'
	@echo '📦  Cleaning previous builds and venvs...'
	@rm -rf .venv
	@cd gen/python && rm -rf dist build *.egg-info .venv
	@echo '📦  Building Python wheel (using Docker CI image)...'
	@docker run --rm -v $(CURDIR):/workspace -w /workspace $(DOCKER_ENV_FLAGS) $(DOCKER_CI_IMAGE) bash -c "cd gen/python && export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0 && uv venv && uv pip install build twine setuptools wheel && uv run python -m build --wheel --installer uv"
	@echo ''
	@echo '✅  Running twine check for validation...'
	@docker run --rm -v $(CURDIR):/workspace -w /workspace $(DOCKER_ENV_FLAGS) $(DOCKER_CI_IMAGE) bash -c "cd gen/python && uv pip install twine && uv run python -m twine check dist/* --strict"
	@echo ''
	@echo '📋  Package metadata (from pyproject.toml):'
	@cd gen/python && grep -A 5 "^\[tool.setuptools.packages.find\]" pyproject.toml
	@echo ''
	@echo '📦  Contents of wheel (first 100 files):'
	@cd gen/python && unzip -l dist/*.whl | head -100
	@echo ''
	@echo '📊  Wheel file size:'
	@cd gen/python && ls -lh dist/*.whl
	@echo ''
	@echo '🧹  Note: build artifacts preserved for inspection in gen/python/'
	@echo '    Run: cd gen/python && rm -rf dist/ build/ *.egg-info to clean up'
	@echo '✅  Python dry run complete!'
	@$(MAKE) sep

.PHONY: dry-run-rust
dry-run-rust: ## Dry-run Rust package (shows what will be published)
	@echo '🦀  Rust Package Dry Run'
	@echo '─────────────────────────────────────────'
	@echo '📋  Files that will be included in crate:'
	@cd gen/rust && cargo package --list --allow-dirty | head -100
	@echo ''
	@echo '📦  Creating package tarball...'
	@cd gen/rust && cargo package --allow-dirty
	@echo ''
	@echo '📊  Package tarball info:'
	@cd gen/rust && ls -lh target/package/flyteidl2-*.crate
	@echo ''
	@echo '📦  Contents of crate tarball (first 50 files):'
	@cd gen/rust && tar -tzf target/package/flyteidl2-*.crate | head -50
	@echo ''
	@echo '✅  Validation: Running cargo build on packaged crate...'
	@cd gen/rust && cargo build --release
	@echo ''
	@echo '🧹  Note: target/package/ directory preserved for inspection'
	@echo '    Run: rm -rf gen/rust/target/package/ to clean up'
	@echo '✅  Rust dry run complete!'
	@$(MAKE) sep

.PHONY: dry-run-all
dry-run-all: dry-run-npm dry-run-python dry-run-rust ## Run dry-run for all packages (TypeScript, Python, Rust)
	@echo '🎉  All package dry runs complete!'
	@echo ''
	@echo 'Summary:'
	@echo '  - TypeScript: gen/ts (npm package @flyteorg/flyteidl2)'
	@echo '  - Python:     gen/python/dist/ (PyPI package flyteidl2)'
	@echo '  - Rust:       gen/rust/target/package/ (crates.io package flyteidl2)'
	@echo ''
	@echo 'Clean up artifacts with:'
	@echo '  - rm -f gen/ts/*.tgz'
	@echo '  - rm -rf gen/python/dist/'
	@echo '  - rm -rf gen/rust/target/package/'
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
