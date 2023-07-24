export REPOSITORY=flytepropeller
include boilerplate/flyte/docker_build/Makefile
include boilerplate/flyte/golang_test_targets/Makefile
include boilerplate/flyte/end2end/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

.PHONY: linux_compile
linux_compile: export CGO_ENABLED ?= 0
linux_compile: export GOOS ?= linux
linux_compile:
	go build -o /artifacts/flytepropeller ./cmd/controller/main.go
	go build -o /artifacts/flytepropeller-manager ./cmd/manager/main.go
	go build -o /artifacts/kubectl-flyte ./cmd/kubectl-flyte/main.go

.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o bin/flytepropeller ./cmd/controller/main.go
	go build -o bin/flytepropeller-manager ./cmd/manager/main.go
	go build -o bin/kubectl-flyte ./cmd/kubectl-flyte/main.go && cp bin/kubectl-flyte ${GOPATH}/bin

cross_compile:
	@glide install
	@mkdir -p ./bin/cross
	go build -o bin/cross/flytepropeller ./cmd/controller/main.go
	go build -o bin/cross/flytepropeller-manager ./cmd/manager/main.go
	go build -o bin/cross/kubectl-flyte ./cmd/kubectl-flyte/main.go

op_code_generate:
	@RESOURCE_NAME=flyteworkflow OPERATOR_PKG=github.com/flyteorg/flytepropeller ./hack/update-codegen.sh

benchmark:
	mkdir -p ./bin/benchmark
	@go test -run=^$ -bench=. -cpuprofile=cpu.out -memprofile=mem.out ./pkg/controller/nodes/. && mv *.out ./bin/benchmark/ && mv *.test ./bin/benchmark/

# server starts the service in development mode
.PHONY: server
server:
	@go run ./cmd/controller/main.go --alsologtostderr --propeller.kube-config=$(HOME)/.kube/config

# manager starts the manager service in development mode
.PHONY: manager
manager:
	@go run ./cmd/manager/main.go --alsologtostderr --propeller.kube-config=$(HOME)/.kube/config

clean:
	rm -rf bin

# Generate golden files. Add test packages that generate golden files here.
golden:
	go test ./cmd/kubectl-flyte/cmd -update
	go test ./pkg/compiler/test -update
