export REPOSITORY=datacatalog
include ../boilerplate/flyte/docker_build/Makefile
include ../boilerplate/flyte/golang_test_targets/Makefile

.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o bin/datacatalog ./cmd/main.go

.PHONY: linux_compile
linux_compile: export CGO_ENABLED ?= 0
linux_compile: export GOOS ?= linux
linux_compile:
	go build -o /artifacts/datacatalog ./cmd/
