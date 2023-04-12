export REPOSITORY=datacatalog
include boilerplate/flyte/docker_build/Makefile
include boilerplate/flyte/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o datacatalog ./cmd/main.go && mv ./datacatalog ./bin

.PHONY: linux_compile
linux_compile: export CGO_ENABLED = 0
linux_compile: export GOOS = linux
linux_compile:
	go build -o /artifacts/datacatalog ./cmd/

.PHONY: generate
generate:
	@go generate ./...
