
.PHONY: linux_compile
linux_compile: export CGO_ENABLED ?= 0
linux_compile: export GOOS ?= linux
linux_compile:
	go build -o /artifacts/flyteadmin -ldflags=$(LD_FLAGS) ./cmd/
