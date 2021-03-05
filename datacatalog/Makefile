export REPOSITORY=datacatalog
include boilerplate/lyft/docker_build/Makefile
include boilerplate/lyft/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o datacatalog ./cmd/main.go && mv ./datacatalog ./bin

.PHONY: linux_compile
linux_compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /artifacts/datacatalog ./cmd/

.PHONY: generate_idl
generate_idl:
	# TODO, move the protos to flyteidl. Currently to generate protos it is
	# assumed that the flyteidl repo is checked out in an adjoining directory.
	# We could use vendoring - but that causes problems when compiling
	protoc -I ../flyteidl/protos/ -I ./protos/idl/. --go_out=plugins=grpc:protos/gen ./protos/idl/service.proto

.PHONY: generate
generate:
	@go generate ./...
