.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo '🆘  Showing help message'
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: buf-dep
buf-dep: ## Update buf modules
	@echo '📦  Updating buf modules'
	buf dep update

.PHONY: buf
buf: buf-dep buf-lint buf-rust buf-python buf-go buf-ts ## Generate all protocol buffer files for all languages
	@echo '🛠️  Finished generating all protocol buffer files for all languages'

.PHONY: buf-lint
buf-lint: ## Lint protocol buffer files
	@echo '🧹  Linting protocol buffer files'
	buf lint

.PHONY: buf-ts
buf-ts: ## Generate TypeScript protocol buffer files
	@echo '🟦  Generating TypeScript protocol buffer files'
	buf generate --clean --template buf.gen.ts.yaml
	@cp flyteidl2/gen_utils/ts/* gen/ts/

.PHONY: buf-go
buf-go: ## Generate Go protocol buffer files
	@echo '🟩  Generating Go protocol buffer files'
	buf generate --clean --template buf.gen.go.yaml

.PHONY: buf-rust
buf-rust: ## Generate Rust protocol buffer files
	@echo '🦀  Generating Rust protocol buffer files'
	buf generate --clean --template buf.gen.rust.yaml
	@cp -R flyteidl2/gen_utils/rust/* gen/rust/
	@cd gen/rust && cargo update --aggressive

export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0
.PHONY: buf-python
buf-python: ## Generate Python protocol buffer files
	@echo '🐍  Generating Python protocol buffer files'
	buf generate --clean --template buf.gen.python.yaml
	@cp flyteidl2/gen_utils/python/* gen/python/
	@find gen/python -type d -exec touch {}/__init__.py \;
	@cd gen/python && uv lock

.PHONY: go_tidy
go_tidy: ## Run go mod tidy
	@echo '🧹  Running go mod tidy'
	go mod tidy

.PHONY: go-tidy
go-tidy: go_tidy ## Run go mod tidy

.PHONY: mocks
mocks: ## Generate go mocks
	@echo "🧪  Generating go mocks"
	mockery

.PHONY: gen
gen: buf go_tidy mocks ## Generates everything in the 'gen' directory
	@echo '⚡  Finished generating everything in the gen directory'
