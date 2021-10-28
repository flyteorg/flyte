export REPOSITORY=flytepropeller
include boilerplate/flyte/docker_build/Makefile
include boilerplate/flyte/golang_test_targets/Makefile
include boilerplate/flyte/end2end/Makefile

CRD_OPTIONS ?= "crd:generateEmbeddedObjectMeta=true"
CONTROLLER_GEN = $(shell pwd)/bin/controller-gen
OPENAPI_GEN = $(shell pwd)/bin/openapi-gen
CRD_FILE =  $(shell pwd)/manifests/base/crds/flyteworkflow.flyte.net_flyteworkflows.yaml

controller-gen: ## Download controller-gen locally if necessary.
	$(call go-get-tool,$(CONTROLLER_GEN),sigs.k8s.io/controller-tools/cmd/controller-gen@v0.7.0)

openapi-gen:
	$(call go-get-tool,$(OPENAPI_GEN),k8s.io/kube-openapi/cmd/openapi-gen@v0.0.0-20200805222855-6aeccd4b50c6)

manifests: controller-gen ## Generate CustomResourceDefinition objects.
	$(CONTROLLER_GEN) $(CRD_OPTIONS) paths="./pkg/apis/flyteworkflow/..." output:crd:artifacts:config=manifests/base/crds || true
	# Kubernetes cannot allow additionalProperties value to be empty.
	# More detail: https://github.com/kubernetes/kubernetes/issues/90504
	perl -p -i -e 's/additionalProperties: {}/additionalProperties: true/g' $(CRD_FILE)
	perl -p -i -e 's/flyteworkflows.flyteworkflow.flyte.net/flyteworkflows.flyte.lyft.com/g' $(CRD_FILE)
	perl -p -i -e 's/flyteworkflow.flyte.net/flyte.lyft.com/g' $(CRD_FILE)

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

.PHONY: linux_compile
linux_compile:
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /artifacts/flytepropeller ./cmd/controller/main.go
	GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /artifacts/kubectl-flyte ./cmd/kubectl-flyte/main.go

.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o bin/flytepropeller ./cmd/controller/main.go
	go build -o bin/kubectl-flyte ./cmd/kubectl-flyte/main.go && cp bin/kubectl-flyte ${GOPATH}/bin

cross_compile:
	@glide install
	@mkdir -p ./bin/cross
	GOOS=linux GOARCH=amd64 go build -o bin/cross/flytepropeller ./cmd/controller/main.go
	GOOS=linux GOARCH=amd64 go build -o bin/cross/kubectl-flyte ./cmd/kubectl-flyte/main.go

openapi_generate: openapi-gen
	$(OPENAPI_GEN) -i github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1 \
				   -p github.com/flyteorg/flytepropeller/pkg/apis/flyteworkflow/v1alpha1 \
				   --go-header-file hack/boilerplate.go.txt 2>&1 > /dev/null

op_code_generate: openapi_generate manifests
	@RESOURCE_NAME=flyteworkflow OPERATOR_PKG=github.com/flyteorg/flytepropeller ./hack/update-codegen.sh

benchmark:
	mkdir -p ./bin/benchmark
	@go test -run=^$ -bench=. -cpuprofile=cpu.out -memprofile=mem.out ./pkg/controller/nodes/. && mv *.out ./bin/benchmark/ && mv *.test ./bin/benchmark/

# server starts the service in development mode
.PHONY: server
server:
	@go run ./cmd/controller/main.go --alsologtostderr --propeller.kube-config=$(HOME)/.kube/config

clean:
	rm -rf bin

# Generate golden files. Add test packages that generate golden files here.
golden:
	go test ./cmd/kubectl-flyte/cmd -update
	go test ./pkg/compiler/test -update

.PHONY: generate
generate: download_tooling
	@go generate ./...

# go-get-tool will 'go get' any package $2 and install it to $1.
PROJECT_DIR := $(shell dirname $(abspath $(firstword $(MAKEFILE_LIST))))
define go-get-tool
@[ -f $(1) ] || { \
set -e ;\
TMP_DIR=$$(mktemp -d) ;\
cd $$TMP_DIR ;\
go mod init tmp ;\
echo "Downloading $(2)" ;\
GOBIN=$(PROJECT_DIR)/bin go get $(2) ;\
rm -rf $$TMP_DIR ;\
}
endef
