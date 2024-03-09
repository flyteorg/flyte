export REPOSITORY=flytestdlib
include ../boilerplate/flyte/golang_test_targets/Makefile

# Generate golden files. Add test packages that generate golden files here.
golden:
	go test ./cli/pflags/api -update
	go test ./config -update
	go test ./storage -update
	go test ./tests -update

# This is the only target that should be overridden by the project. Get your binary into ${GOREPO}/bin
.PHONY: compile
compile:
	mkdir -p ./bin
	go build -o ./bin/pflags ./cli/pflags/main.go
