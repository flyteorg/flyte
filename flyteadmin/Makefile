export REPOSITORY=flyteadmin
export FLYTE_SCHEDULER_REPOSITORY=flytescheduler
export REPO_ROOT=..
include ../boilerplate/flyte/docker_build/Makefile
include ../boilerplate/flyte/golang_test_targets/Makefile
include ../boilerplate/flyte/end2end/Makefile

GIT_VERSION := $(shell git describe --tags --long --match "v*" --first-parent)
GIT_HASH := $(shell git rev-parse --short HEAD)
TIMESTAMP := $(shell date '+%Y-%m-%d')
# TODO(monorepo): Do we need to change this? This is used in the service that provides a version.
PACKAGE ?=github.com/flyteorg/flyte/flytestdlib

LD_FLAGS="-s -w -X $(PACKAGE)/version.Version=$(GIT_VERSION) -X $(PACKAGE)/version.Build=$(GIT_HASH) -X $(PACKAGE)/version.BuildTime=$(TIMESTAMP)"

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

.PHONY: docker_build_scheduler
docker_build_scheduler:
	docker build -t $$FLYTE_SCHEDULER_REPOSITORY:$(GIT_HASH) -f Dockerfile.scheduler .

.PHONY: integration
integration: export CGO_ENABLED ?= 0
integration: export GOFLAGS ?= -count=1
integration:
	go test -v -tags=integration ./tests/...

.PHONY: k8s_integration
k8s_integration:
	@script/integration/launch.sh

.PHONY: k8s_integration_execute
k8s_integration_execute:
	@script/integration/execute.sh

.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o bin/flyteadmin -ldflags=$(LD_FLAGS) ./cmd/

.PHONY: compile_debug
compile_debug:
	go build -o flyteadmin -gcflags='all=-N -l' ./cmd/ && mv ./flyteadmin ${GOPATH}/bin


.PHONY: compile_scheduler
compile_scheduler:
	mkdir -p ./bin
	go build -o bin/flytescheduler -ldflags=$(LD_FLAGS) ./cmd/scheduler/

.PHONY: compile_scheduler_debug
compile_scheduler_debug:
	go build -o flytescheduler -gcflags='all=-N -l' ./cmd/scheduler/ && mv ./flytescheduler ${GOPATH}/bin


.PHONY: linux_compile
linux_compile: export CGO_ENABLED ?= 0
linux_compile: export GOOS ?= linux
linux_compile:
	go build -o /artifacts/flyteadmin -ldflags=$(LD_FLAGS) ./cmd/

.PHONY: linux_compile_scheduler
linux_compile_scheduler: export CGO_ENABLED ?= 0
linux_compile_scheduler: export GOOS ?= linux
linux_compile_scheduler:
	go build -o /artifacts/flytescheduler -ldflags=$(LD_FLAGS) ./cmd/scheduler/

.PHONY: server
server:
	go run cmd/main.go serve  --server.kube-config ~/.kube/config  --config flyteadmin_config.yaml

.PHONY: scheduler
scheduler:
	go run cmd/scheduler/main.go run  --server.kube-config ~/.kube/config  --config flyteadmin_config.yaml

.PHONY: migrate
migrate:
	go run cmd/main.go migrate run --server.kube-config ~/.kube/config  --config flyteadmin_config.yaml

.PHONY: seed_projects
seed_projects:
	go run cmd/main.go migrate seed-projects project admintests flytekit --server.kube-config ~/.kube/config  --config flyteadmin_config.yaml

all: compile
