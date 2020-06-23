export REPOSITORY=flyteplugins
include boilerplate/lyft/docker_build/Makefile
include boilerplate/lyft/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

generate: download_tooling
	@go generate ./...

clean:
	rm -rf bin

.PHONY: linux_compile
linux_compile:
	cd copilot; GOOS=linux GOARCH=amd64 CGO_ENABLED=0 go build -o /artifacts/flyte-copilot .; cd -

.PHONY: compile
compile:
	mkdir -p ./artifacts
	cd copilot; go build -o ../artifacts/flyte-copilot .; cd -

cross_compile:
	@glide install
	@mkdir -p ./artifacts/cross
	cd copilot; GOOS=linux GOARCH=amd64 go build -o ../artifacts/flyte-copilot .; cd -
