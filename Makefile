.PHONY: buf
buf: buf-rust buf-python buf-go buf-ts

.PHONY: buf-ts
buf-ts:
	buf generate --clean --template buf.gen.ts.yaml
	cp idl/gen_utils/ts/* gen/ts/

.PHONY: buf-go
buf-go:
	buf generate --clean --template buf.gen.go.yaml

.PHONY: buf-rust
buf-rust:
	buf generate --clean --template buf.gen.rust.yaml
	cp idl/gen_utils/rust/* gen/rust/src/
	cd gen/rust && cargo update --aggressive

.PHONY: buf-python
buf-python:
	buf generate --clean --template buf.gen.python.yaml
	cp idl/gen_utils/python/* gen/python/
	cd gen/python && uv lock

.PHONY: generate
generate: buf