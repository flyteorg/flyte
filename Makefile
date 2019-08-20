export REPOSITORY=flyteplugins
include boilerplate/lyft/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

generate:
	which pflags || (go get github.com/lyft/flytestdlib/cli/pflags)
	which mockery || (go get github.com/vektra/mockery/cmd/mockery)
	@go generate ./...

clean:
	rm -rf bin
