export REPOSITORY=flytectl
include boilerplate/lyft/golang_test_targets/Makefile

generate:
	go test github.com/lyft/flytectl/cmd --update

compile:
	go build -o bin/flytectl main.go

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh
