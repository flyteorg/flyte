include boilerplate/flyte/golang_test_targets/Makefile

.PHONY: linux_compile
linux_compile: export CGO_ENABLED ?= 0
linux_compile: export GOOS ?= linux
linux_compile:
	go build -o /artifacts/flyteartifacts -ldflags=$(LD_FLAGS) ./cmd/
