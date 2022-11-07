export REPOSITORY=flytestdlib
include boilerplate/flyte/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@curl https://raw.githubusercontent.com/flyteorg/boilerplate/master/boilerplate/update.sh -o boilerplate/update.sh
	@boilerplate/update.sh

# Generate golden files. Add test packages that generate golden files here.
golden:
	go test ./cli/pflags/api -update
	go test ./config -update
	go test ./storage -update
	go test ./tests -update


generate:
	@echo "************************ go generate **********************************"
	go generate ./...

# This is the only target that should be overriden by the project. Get your binary into ${GOREPO}/bin
.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o pflags ./cli/pflags/main.go && mv ./pflags ./bin


.PHONY: test_unit_codecov
test_unit_codecov:
	go test ./... -race -coverprofile=coverage.txt -covermode=atomic
