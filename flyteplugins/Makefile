export GO111MODULE=off
export REPOSITORY=flyteplugins
include boilerplate/lyft/golang_test_targets/Makefile

.PHONY: update_boilerplate
update_boilerplate:
	@boilerplate/update.sh

generate:
	which pflags || (go get github.com/lyft/flytestdlib/cli/pflags)
	which mockery || (go install github.com/lyft/flyteidl/vendor/github.com/vektra/mockery/cmd/mockery)
	which enumer || (go get github.com/alvaroloes/enumer)
	@go generate ./...

clean:
	rm -rf bin
