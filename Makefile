.PHONY: buf
buf: buf-rust buf-python buf-go buf-ts

.PHONY: buf-ts
buf-ts:
	buf generate --clean --template buf.gen.ts.yaml
	cp idl2/gen_utils/ts/* gen/ts/

.PHONY: buf-go
buf-go:
	buf generate --clean --template buf.gen.go.yaml

.PHONY: buf-rust
buf-rust:
	buf generate --clean --template buf.gen.rust.yaml
	cp idl2/gen_utils/rust/* gen/rust/src/
	cd gen/rust && cargo update --aggressive

export SETUPTOOLS_SCM_PRETEND_VERSION=0.0.0
.PHONY: buf-python
buf-python:
	buf generate --clean --template buf.gen.python.yaml
	cp idl2/gen_utils/python/* gen/python/
	cd gen/python && uv lock

.PHONY: generate
generate: buf