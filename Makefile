.DEFAULT_GOAL := help

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

.PHONY: buf
buf: buf-dep buf-lint buf-rust buf-python buf-go buf-ts ## Generate all protocol buffer files for all languages
	@echo '🛠️  Finished generating all protocol buffer files for all languages'
	@$(MAKE) sep

.PHONY: buf-lint
buf-lint: ## Lint protocol buffer files
	@echo '🧹  Linting protocol buffer files'
	buf lint $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: buf-ts
buf-ts: ## Generate TypeScript protocol buffer files
	@echo '🟦  Generating TypeScript protocol buffer files'
	buf generate --clean --template buf.gen.ts.yaml $(OUT_REDIRECT)
	@cp flyteidl2/gen_utils/ts/* gen/ts/
	@$(MAKE) sep

.PHONY: buf-go
buf-go: ## Generate Go protocol buffer files
	@echo '🟩  Generating Go protocol buffer files'
	buf generate --clean --template buf.gen.go.yaml $(OUT_REDIRECT)
	@$(MAKE) sep

.PHONY: buf-rust
buf-rust: ## Generate Rust protocol buffer files
	@echo '🦀  Generating Rust protocol buffer files'
	buf generate --clean --template buf.gen.rust.yaml $(OUT_REDIRECT)
	@cp -R flyteidl2/gen_utils/rust/* gen/rust/
	@cd gen/rust && cargo update --aggressive
	@$(MAKE) sep

export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0
.PHONY: buf-python
buf-python: ## Generate Python protocol buffer files
	@echo '🐍  Generating Python protocol buffer files'
	buf generate --clean --template buf.gen.python.yaml $(OUT_REDIRECT)
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

