export REPOSITORY=flyteadmin
include boilerplate/flyte/docker_build/Makefile
include boilerplate/flyte/golang_test_targets/Makefile
include boilerplate/flyte/end2end/Makefile

GIT_VERSION := $(shell git describe --always --tags)
GIT_HASH := $(shell git rev-parse --short HEAD)
TIMESTAMP := $(shell date '+%Y-%m-%d')
PACKAGE ?=github.com/flyteorg/flytestdlib

LD_FLAGS="-s -w -X $(PACKAGE)/version.Version=$(GIT_VERSION) -X $(PACKAGE)/version.Build=$(GIT_HASH) -X $(PACKAGE)/version.BuildTime=$(TIMESTAMP)"

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

.PHONY: integration
integration:
	CGO_ENABLED=0 GOFLAGS="-count=1" go test -v -tags=integration ./tests/...

.PHONY: k8s_integration
k8s_integration:
	@script/integration/launch.sh

.PHONY: k8s_integration_execute
k8s_integration_execute:
	@script/integration/execute.sh

.PHONY: compile
compile:
	go build -o flyteadmin -ldflags=$(LD_FLAGS) ./cmd/ && mv ./flyteadmin ${GOPATH}/bin

.PHONY: linux_compile
linux_compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0  go build -o /artifacts/flyteadmin -ldflags=$(LD_FLAGS) ./cmd/

.PHONY: server
server:
	go run cmd/main.go serve  --server.kube-config ~/.kube/config  --config flyteadmin_config.yaml

.PHONY: migrate
migrate:
	go run cmd/main.go migrate run --server.kube-config ~/.kube/config  --config flyteadmin_config.yaml

.PHONY: seed_projects
seed_projects:
	go run cmd/main.go migrate seed-projects project admintests flytekit --server.kube-config ~/.kube/config  --config flyteadmin_config.yaml

all: compile

generate: download_tooling
	@go generate ./...
