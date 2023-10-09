export REPOSITORY=flytecopilot
include boilerplate/flyte/docker_build/Makefile
include boilerplate/flyte/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

clean:
	rm -rf bin

.PHONY: linux_compile
linux_compile: export CGO_ENABLED ?= 0
linux_compile: export GOOS ?= linux
linux_compile:
	go build -o /artifacts/flyte-copilot .

.PHONY: compile
compile:
	mkdir -p ./artifacts
	go build -o ./artifacts/flyte-copilot .

cross_compile:
	@mkdir -p ./artifacts/cross
	GOOS=linux GOARCH=amd64 go build -o ./artifacts/flyte-copilot .
