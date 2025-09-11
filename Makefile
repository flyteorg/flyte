.DEFAULT_GOAL := help

.PHONY: help
help: ## Show this help message
	@echo 'Usage: make [target]'
	@echo ''
	@echo 'Available targets:'
	@awk 'BEGIN {FS = ":.*?## "} /^[a-zA-Z_-]+:.*?## / {printf "  %-15s %s\n", $$1, $$2}' $(MAKEFILE_LIST)

.PHONY: buf
buf: buf-rust buf-python buf-go buf-ts ## Generate all protocol buffer files for all languages

.PHONY: buf-ts
buf-ts: ## Generate TypeScript protocol buffer files
	buf generate --clean --template buf.gen.ts.yaml
	cp flyteidl2/gen_utils/ts/* gen/ts/

.PHONY: buf-go
buf-go: ## Generate Go protocol buffer files
	buf generate --clean --template buf.gen.go.yaml

.PHONY: buf-rust
buf-rust: ## Generate Rust protocol buffer files
	buf generate --clean --template buf.gen.rust.yaml
	cp flyteidl2/gen_utils/rust/* gen/rust/src/
	cd gen/rust && cargo update --aggressive

export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0
.PHONY: buf-python
buf-python: ## Generate Python protocol buffer files
	buf generate --clean --template buf.gen.python.yaml
	buf generate buf.build/bufbuild/protovalidate --template buf.gen.python.yaml
	cp flyteidl2/gen_utils/python/* gen/python/
	find gen/python -type d -exec touch {}/__init__.py \;
	cd gen/python && uv lock

.PHONY: gen
gen: buf ## Generates everything in the 'gen' directory
